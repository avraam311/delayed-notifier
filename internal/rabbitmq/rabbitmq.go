package rabbitmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/rabbitmq"
)

type RabbitMq struct {
	publisher *rabbitmq.Publisher
	consumer  *rabbitmq.Consumer
	ch        *amqp.Channel
	cfg       *config.Config
}

func New(cfg *config.Config) (*RabbitMq, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.GetString("rabbitmq.user"),
		cfg.GetString("rabbitmq.password"),
		cfg.GetString("rabbitmq.host"),
		cfg.GetString("rabbitmq.port"))
	conn, err := rabbitmq.Connect(url, cfg.GetInt("rabbitmq.retries"), time.Second*10)
	if err != nil {
		return nil, err
	}

	rabbitMqChan, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	notExchange := cfg.GetString("rabbitmq.not_exchange")
	mainExchange := rabbitmq.NewExchange(
		notExchange,
		"x-delayed-message",
	)
	args := amqp.Table{
		"x-delayed-type": "direct",
	}
	mainExchange.Args = args
	mainExchange.Durable = true

	err = mainExchange.BindToChannel(rabbitMqChan)
	if err != nil {
		return nil, err
	}

	retryNotExchange := cfg.GetString("rabbitmq.retry_not_exchange")
	retryExchange := rabbitmq.NewExchange(
		retryNotExchange,
		"direct",
	)
	retryExchange.Durable = true

	args = amqp.Table{
		"x-dead-letter-exchange": retryNotExchange,
	}

	err = retryExchange.BindToChannel(rabbitMqChan)
	if err != nil {
		return nil, err
	}

	mainQueueManager := rabbitmq.NewQueueManager(rabbitMqChan)
	queueCfg := rabbitmq.QueueConfig{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       args,
	}
	notQueue := cfg.GetString("rabbitmq.not_queue")
	mainQueue, err := mainQueueManager.DeclareQueue(
		notQueue,
		queueCfg,
	)
	if err != nil {
		return nil, err
	}

	routingKey := cfg.GetString("rabbitmq.routing_key")
	err = rabbitMqChan.QueueBind(
		mainQueue.Name,
		routingKey,
		notExchange,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	retryArgs := amqp.Table{
		"x-message-ttl":          int32(cfg.GetInt("rabbitmq.ttl")),
		"x-dead-letter-exchange": notExchange,
	}

	retryQueueManager := rabbitmq.NewQueueManager(rabbitMqChan)
	retryQueueCfg := rabbitmq.QueueConfig{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       retryArgs,
	}
	notQueueRetry := cfg.GetString("rabbitmq.not_queue_retry")
	retryQueue, err := retryQueueManager.DeclareQueue(
		notQueueRetry,
		retryQueueCfg,
	)
	if err != nil {
		return nil, err
	}

	err = rabbitMqChan.QueueBind(
		retryQueue.Name,
		routingKey,
		retryNotExchange,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	consumerConfig := rabbitmq.NewConsumerConfig(notQueue)

	publisher := rabbitmq.NewPublisher(rabbitMqChan, notExchange)
	consumer := rabbitmq.NewConsumer(rabbitMqChan, consumerConfig)

	return &RabbitMq{
		publisher: publisher,
		consumer:  consumer,
		ch:        rabbitMqChan,
		cfg:       cfg,
	}, nil
}

func (r *RabbitMq) Publish(routingKey string, message []byte, contentType string, delay time.Duration) error {
	headers := amqp.Table{
		"x-delay": int64(delay / time.Millisecond),
	}
	publishinOptions := rabbitmq.PublishingOptions{
		Headers: headers,
	}

	err := r.publisher.Publish(message, routingKey, contentType, publishinOptions)
	if err != nil {
		return err
	}
	return nil
}

func (r *RabbitMq) Consume(msgChan chan []byte) error {
	err := r.consumer.Consume(msgChan)
	if err != nil {
		return err
	}
	return nil
}

func (r *RabbitMq) Close() error {
	if err := r.ch.Close(); err != nil {
		return err
	}
	return nil
}

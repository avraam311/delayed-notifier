package rabbitmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
)

type RabbitMq struct {
	publisher *rabbitmq.Publisher
	consumer  *rabbitmq.Consumer
}

func New() (*RabbitMq, error) {
	conn, err := rabbitmq.Connect("amqp://guest:guest@rabbitmq:5672/", 3, time.Second*10)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq/rabbitmq.go - failed to connect to RabbitMq - %w", err)
	}

	rabbitMqChan, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq/rabbitmq.go - failed to open channel - %w", err)
	}

	mainExchange := rabbitmq.NewExchange(
		"notifications-exchange",
		"x-delayed-message",
	)
	args := amqp.Table{
		"x-delayed-type": "direct",
	}
	mainExchange.Args = args
	mainExchange.Durable = true

	err = mainExchange.BindToChannel(rabbitMqChan)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq/rabbitmq.go - failed to declare exchange - %w", err)
	}

	retryExchange := rabbitmq.NewExchange(
		"retry-notifications-exchange",
		"direct",
	)
	retryExchange.Durable = true

	args = amqp.Table{
		"x-dead-letter-exchange": "retry-notifications-exchange",
	}

	err = retryExchange.BindToChannel(rabbitMqChan)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq/rabbitmq.go - failed to declare retry exchange - %w", err)
	}

	mainQueueManager := rabbitmq.NewQueueManager(rabbitMqChan)
	queueCfg := rabbitmq.QueueConfig{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       args,
	}
	mainQueue, err := mainQueueManager.DeclareQueue(
		"notifications-queue",
		queueCfg,
	)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq/rabbitmq.go - failed to declare queue - %w", err)
	}

	err = rabbitMqChan.QueueBind(
		mainQueue.Name,
		"notifications-key",
		"notifications-exchange",
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq/rabbitmq.go - failed to bind - %w", err)
	}

	retryArgs := amqp.Table{
		"x-message-ttl":          int32(60000),
		"x-dead-letter-exchange": "notifications-exchange",
	}

	retryQueueManager := rabbitmq.NewQueueManager(rabbitMqChan)
	retryQueueCfg := rabbitmq.QueueConfig{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       retryArgs,
	}
	retryQueue, err := retryQueueManager.DeclareQueue(
		"retry-notifications-queue",
		retryQueueCfg,
	)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq/rabbitmq.go - failed to declare retry queue - %w", err)
	}

	err = rabbitMqChan.QueueBind(
		retryQueue.Name,
		"notifications-key",
		"retry-notifications-exchange",
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq/rabbitmq.go - failed to bind - %w", err)
	}

	consumerConfig := rabbitmq.NewConsumerConfig("notifications-queue")

	publisher := rabbitmq.NewPublisher(rabbitMqChan, "notifications-exchange")
	consumer := rabbitmq.NewConsumer(rabbitMqChan, consumerConfig)

	return &RabbitMq{
		publisher: publisher,
		consumer:  consumer,
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
		return fmt.Errorf("rabbitmq/publisher.go - failed to publish message - %w", err)
	}
	return nil
}

func (r *RabbitMq) Consume(msgChan chan []byte) error {
	err := r.consumer.Consume(msgChan)
	if err != nil {
		return fmt.Errorf("rabbitmq/consumer.go - failed to consume message - %w", err)
	}
	return nil
}

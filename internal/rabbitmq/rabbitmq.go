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

	e := rabbitmq.NewExchange(
		"notifications-exchange",
		"x-delayed-message",
	)
	args := amqp.Table{
		"x-delayed-type": "direct",
	}
	e.Args = args
	e.Durable = true
	err = e.BindToChannel(rabbitMqChan)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq/rabbitmq.go - failed to declare exchange - %w", err)
	}

	qm := rabbitmq.NewQueueManager(rabbitMqChan)
	queueCfg := rabbitmq.QueueConfig{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
	}
	queue, err := qm.DeclareQueue(
		"notifications-queue",
		queueCfg,
	)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq/rabbitmq.go - failed to declare queue - %w", err)
	}

	err = rabbitMqChan.QueueBind(
		queue.Name,
		"notifications-key",
		"notifications-exchange",
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

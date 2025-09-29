package main

import (
	"time"

	"github.com/avraam311/delayed-notifier/internal/rabbitmq"
	"github.com/avraam311/delayed-notifier/internal/worker"
	"github.com/wb-go/wbf/zlog"
)

const (
	workersCount = 5
)

func main() {
	zlog.Init()
	RMQ, err := rabbitmq.New()
	if err != nil {
		zlog.Logger.Error().Err(err)
		panic("failed to init rabbitMq")
	}
	RMQ.Publish("notifications-key", []byte{'a'}, "string", time.Second*1)
	RMQ.Publish("notifications-key", []byte{'b'}, "string", time.Second*3)
	RMQ.Publish("notifications-key", []byte{'c'}, "string", time.Second*5)
	work := worker.New(RMQ, workersCount)
	zlog.Logger.Info().Msg("worker is running")
	work.Run()
}

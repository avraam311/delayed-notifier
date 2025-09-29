package main

import (
	"encoding/json"
	"time"

	"github.com/avraam311/delayed-notifier/internal/config"
	"github.com/avraam311/delayed-notifier/internal/models"
	"github.com/avraam311/delayed-notifier/internal/rabbitmq"
	"github.com/avraam311/delayed-notifier/internal/sender"
	"github.com/avraam311/delayed-notifier/internal/worker"

	"github.com/wb-go/wbf/zlog"
)

const (
	workersCount = 5
)

func main() {
	zlog.Init()
	cfg, err := config.MustLoad()
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to init config")
	}
	RMQ, err := rabbitmq.New()
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to init rabbitMq")
	}

	not := models.Notification{
		UserID:  6176317973,
		Message: "hi",
	}
	msg, err := json.Marshal(not)
	if err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to marshal into json")
	}

	RMQ.Publish("notifications-key", msg, "application/json", time.Second*1)
	RMQ.Publish("notifications-key", msg, "application/json", time.Second*20)
	RMQ.Publish("notifications-key", msg, "application/json", time.Second*40)

	tgBot, err := sender.New(cfg.Env.BotToken)
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to init tgBot")
	}

	work := worker.New(RMQ, workersCount, tgBot)
	zlog.Logger.Info().Msg("worker is running")
	work.Run()
}

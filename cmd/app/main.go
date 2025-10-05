package main

import (
	"encoding/json"
	"time"

	"github.com/avraam311/delayed-notifier/internal/config"
	"github.com/avraam311/delayed-notifier/internal/models/domain"
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

	dateTime, err := time.Parse(time.DateTime, "2025-10-3 18:30:00")
	not := domain.Notification{
		Message:  "hi",
		DateTime: dateTime,
		Mail:     "example@mail.ru",
		TgID:     617,
	}
	msg, err := json.Marshal(not)
	if err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to marshal into json")
	}

	RMQ.Publish("notifications-key", msg, "application/json", time.Second*1)
	RMQ.Publish("notifications-key", msg, "application/json", time.Second*20)
	RMQ.Publish("notifications-key", msg, "application/json", time.Second*40)

	tgBot, err := sender.NewBot(cfg.Env.BotToken)
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to init tgBot")
	}
	Mail := sender.NewMail("smtp.mail.ru", "587", cfg.Env.MailLogin, cfg.Env.MailPassword)

	work := worker.New(RMQ, workersCount, tgBot, Mail)
	zlog.Logger.Info().Msg("worker is running")
	work.Run()
}

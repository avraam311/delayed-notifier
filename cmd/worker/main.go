package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/avraam311/delayed-notifier/internal/rabbitmq"
	"github.com/avraam311/delayed-notifier/internal/sender"
	"github.com/avraam311/delayed-notifier/internal/worker"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/zlog"
)

const (
	configFilePath = "./config/local.yaml"
	envFilePath    = "./.env"
	envPrefix      = ""
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	zlog.Init()
	cfg := config.New()
	err := cfg.Load(configFilePath, envFilePath, envPrefix)
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to initialize config")
	}

	rMQ, err := rabbitmq.New(cfg)
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to initialize rabbitMQ")
	}

	tgBot, err := sender.NewBot(cfg.GetString("TELEGRAM_TOKEN"))
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to initialize tg bot")
	}

	mail := sender.NewMail(cfg.GetString("SMTP_HOST"),
		cfg.GetString("SMTP_PORT"),
		cfg.GetString("SMTP_USER"),
		cfg.GetString("SMTP_FROM"),
		cfg.GetString("SMTP_PASSWORD"),
	)

	work := worker.New(rMQ, cfg.GetInt("workers.count"), tgBot, mail, cfg)
	go work.Run(ctx)
	zlog.Logger.Info().Msg("worker is running")

	<-ctx.Done()

	if err := rMQ.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close RabbitMQ channel")
	}
}

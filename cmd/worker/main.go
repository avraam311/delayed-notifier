package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/avraam311/delayed-notifier/internal/config"
	"github.com/avraam311/delayed-notifier/internal/rabbitmq"
	"github.com/avraam311/delayed-notifier/internal/sender"
	"github.com/avraam311/delayed-notifier/internal/worker"

	"github.com/wb-go/wbf/zlog"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	zlog.Init()
	cfg, err := config.MustLoad()
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to initialize config")
	}

	rMQ, err := rabbitmq.New()
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to initialize rabbitMQ")
	}

	tgBot, err := sender.NewBot(cfg.Env.BotToken)
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to initialize tg bot")
	}
	mail := sender.NewMail(cfg.Cfg.GetString("host"),
		cfg.Cfg.GetString("port"),
		cfg.Cfg.GetString("from"),
		cfg.Cfg.GetString("password"))

	work := worker.New(rMQ, cfg.Cfg.GetInt("workers"), tgBot, mail)
	go work.Run()

	<-ctx.Done()

	if err := rMQ.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close RabbitMQ channel")
	}
}

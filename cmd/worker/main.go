package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/avraam311/delayed-notifier/internal/rabbitmq"
	notRepo "github.com/avraam311/delayed-notifier/internal/repository/notifications"
	"github.com/avraam311/delayed-notifier/internal/sender"
	"github.com/avraam311/delayed-notifier/internal/worker"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"
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

	opts := &dbpg.Options{
		MaxOpenConns:    cfg.GetInt("db.max_open_conns"),
		MaxIdleConns:    cfg.GetInt("db.max_idle_conns"),
		ConnMaxLifetime: cfg.GetDuration("db.conn_max_lifetime"),
	}
	slavesDNSs := []string{}
	masterDNS := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.GetString("DB_USER"), cfg.GetString("DB_PASSWORD"),
		cfg.GetString("DB_HOST"), cfg.GetString("PORT"),
		cfg.GetString("DB_NAME"), cfg.GetString("DB_SSL"),
	)
	db, err := dbpg.New(masterDNS, slavesDNSs, opts)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	repo := notRepo.NewRepository(db)

	work := worker.New(rMQ, cfg.GetInt("workers.count"), tgBot, mail, cfg, repo)
	go work.Run(ctx)
	zlog.Logger.Info().Msg("worker is running")

	<-ctx.Done()

	if err := rMQ.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close RabbitMQ channel")
	}
}

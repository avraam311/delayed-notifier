package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/avraam311/delayed-notifier/internal/api/handlers"
	"github.com/avraam311/delayed-notifier/internal/api/server"
	"github.com/avraam311/delayed-notifier/internal/rabbitmq"
	notRepo "github.com/avraam311/delayed-notifier/internal/repository/notifications"
	notService "github.com/avraam311/delayed-notifier/internal/service/notifications"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"

	"github.com/go-playground/validator/v10"
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

	val := validator.New()

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

	rMQ, err := rabbitmq.New(cfg)
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to initialize rabbitMQ")
	}

	repo := notRepo.NewRepository(db)
	s := notService.NewService(repo, rMQ, cfg)
	notHandler := handlers.NewHandler(s, val)

	ginMode := cfg.GetString("server.gin_mode")
	r := server.NewRouter(notHandler, ginMode)
	srv := server.NewServer(cfg.GetString("server.port"), r)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			zlog.Logger.Panic().Err(err).Msg("failed to run server")
		}
	}()
	zlog.Logger.Info().Msg("server is running")

	<-ctx.Done()
	zlog.Logger.Info().Msg("graceful shutdown signal recieved")

	shutdownCtx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to shutdown server")
	}

	if err := db.Master.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close master db")
	}
	for i, s := range db.Slaves {
		if err := s.Close(); err != nil {
			zlog.Logger.Error().Err(err).Int("slave number", i).Msg("failed to close slave db")
		}
	}

	if err := rMQ.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to close RabbitMQ channel")
	}
}

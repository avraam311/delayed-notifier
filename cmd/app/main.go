package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/avraam311/delayed-notifier/internal/api/handlers"
	"github.com/avraam311/delayed-notifier/internal/api/server"
	"github.com/avraam311/delayed-notifier/internal/config"
	"github.com/avraam311/delayed-notifier/internal/rabbitmq"
	notRepo "github.com/avraam311/delayed-notifier/internal/repository/notifications"
	notService "github.com/avraam311/delayed-notifier/internal/service/notifications"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"

	"github.com/go-playground/validator/v10"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	zlog.Init()
	cfg, err := config.MustLoad()
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to initialize config")
	}

	val := validator.New()
	rMQ, err := rabbitmq.New(cfg)
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to initialize rabbitMQ")
	}

	opts := &dbpg.Options{
		MaxOpenConns:    cfg.Cfg.GetInt("database.max_open_conns"),
		MaxIdleConns:    cfg.Cfg.GetInt("database.max_idle_conns"),
		ConnMaxLifetime: time.Duration(cfg.Cfg.GetInt("database.conn_max_lifetime")),
	}

	slaveDNSs := make([]string, 0, len(cfg.Cfg.GetSlice("database.slaves")))

	for _, s := range cfg.Cfg.GetSlice("database.slaves") {
		slaveDNSs = append(slaveDNSs, s.DSN())
	}

	db, err := dbpg.New(cfg.Cfg.GetString("database.master"), slaveDNSs, opts)
	if err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to initialize db")
	}

	repo := notRepo.NewRepository(db)
	s := notService.NewService(repo, rMQ)
	notHandler := handlers.NewHandler(s, val)

	r := server.NewRouter(notHandler)
	srv := server.NewServer(cfg.Cfg.GetString("server.port"), r)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			zlog.Logger.Panic().Err(err).Msg("failed to run server")
		}
	}()

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

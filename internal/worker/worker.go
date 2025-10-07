package worker

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/avraam311/delayed-notifier/internal/rabbitmq"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type tgBotI interface {
	SendMessage(msg []byte) error
}

type mailI interface {
	SendMessage(msg []byte) error
}

type Worker struct {
	RMQ          *rabbitmq.RabbitMq
	WorkersCount int
	TgBot        tgBotI
	Mail         mailI
	cfg          *config.Config
}

func New(rMQ *rabbitmq.RabbitMq, workersCount int, tgBot tgBotI, mail mailI, cfg *config.Config) *Worker {
	return &Worker{
		RMQ:          rMQ,
		WorkersCount: workersCount,
		TgBot:        tgBot,
		Mail:         mail,
		cfg:          cfg,
	}
}

func (w *Worker) Run(ctx context.Context) {
	readerCh := make(chan []byte)
	var wg sync.WaitGroup

	go func() {
		err := w.RMQ.Consume(readerCh)
		if err != nil {
			zlog.Logger.Panic().Err(err).Msg("worker/worker.go - failed to start consuming messages")
		}
	}()

	retryStrategy := retry.Strategy{
		Attempts: w.cfg.GetInt("retry.attempts"),
		Delay:    w.cfg.GetDuration("retry.delay"),
		Backoff:  w.cfg.GetFloat64("retry.backoff"),
	}

	wg.Add(w.WorkersCount)
	for i := 0; i < w.WorkersCount; i++ {
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-readerCh:
					err := retry.Do(func() error {
						return w.sendToTg(msg)
					}, retryStrategy)
					if err != nil {
						zlog.Logger.Warn().Err(err).Str("worker_id", strconv.Itoa(id)).Msg("worker/workeg.go - failed to send to tg")
					}
					err = retry.Do(func() error {
						return w.sendToMail(msg)
					}, retryStrategy)
					if err != nil {
						zlog.Logger.Warn().Err(err).Str("worker_id", strconv.Itoa(id)).Msg("worker/workeg.go - failed to send to mail")
					}
				}
			}
		}(i)
	}

	wg.Wait()
}

func (w *Worker) sendToTg(msg []byte) error {
	tgErr := w.TgBot.SendMessage(msg)

	if tgErr != nil {
		return fmt.Errorf("failed to send notification to tg - %w", tgErr)
	}

	return nil
}

func (w *Worker) sendToMail(msg []byte) error {
	mailErr := w.Mail.SendMessage(msg)

	if mailErr != nil {
		return fmt.Errorf("failed to send notification to mail - %w", mailErr)
	}

	return nil
}

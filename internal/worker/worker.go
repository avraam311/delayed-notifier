package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/avraam311/delayed-notifier/internal/rabbitmq"

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
}

func New(rMQ *rabbitmq.RabbitMq, workersCount int, tgBot tgBotI, mail mailI) *Worker {
	return &Worker{
		RMQ:          rMQ,
		WorkersCount: workersCount,
		TgBot:        tgBot,
		Mail:         mail,
	}
}

func (w *Worker) Run() {
	readerCh := make(chan []byte)
	var wg sync.WaitGroup

	go func() {
		err := w.RMQ.Consume(readerCh)
		if err != nil {
			zlog.Logger.Warn().Err(err).Msg("worker/worker.go - failed to consume message")
		}
	}()

	retryStrategy := retry.Strategy{
		Attempts: 3,
		Delay:    time.Minute,
		Backoff:  3,
	}

	wg.Add(w.WorkersCount)
	for i := 0; i < w.WorkersCount; i++ {
		go func() {
			defer wg.Done()
			for msg := range readerCh {
				err := retry.Do(func() error {
					return w.handlerMessage(msg)
				}, retryStrategy)
				if err != nil {
					zlog.Logger.Warn().Err(err).Msg("worker/workeg.go - failed to handle message perfectly")
				}
			}
		}()
	}

	wg.Wait()
}

func (w *Worker) handlerMessage(msg []byte) error {
	tgErr := w.TgBot.SendMessage(msg)

	mailErr := w.Mail.SendMessage(msg)

	if tgErr != nil && mailErr != nil {
		return fmt.Errorf("failed to send notification to both tg and mail")
	} else if tgErr != nil {
		return fmt.Errorf("failed to send notification to tg")
	} else if mailErr != nil {
		return fmt.Errorf("failed to send notification to mail")
	}

	return nil
}

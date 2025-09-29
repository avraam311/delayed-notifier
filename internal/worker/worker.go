package worker

import (
	"sync"

	"github.com/avraam311/delayed-notifier/internal/rabbitmq"

	"github.com/wb-go/wbf/zlog"
)

type tgBot interface {
	SendMessage(msg []byte) error
}

type Worker struct {
	RMQ          *rabbitmq.RabbitMq
	WorkersCount int
	TgBot        tgBot
}

func New(rMQ *rabbitmq.RabbitMq, workersCount int, tgBot tgBot) *Worker {
	return &Worker{
		RMQ:          rMQ,
		WorkersCount: workersCount,
		TgBot:        tgBot,
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

	wg.Add(w.WorkersCount)
	for i := 0; i < w.WorkersCount; i++ {
		go func() {
			defer wg.Done()
			for msg := range readerCh {
				err := w.handlerMessage(msg)
				if err != nil {
					zlog.Logger.Warn().Err(err).Msg("worker/workeg.go - failed to handle message")
				}
			}
		}()
	}

	wg.Wait()
}

func (w *Worker) handlerMessage(msg []byte) error {
	err := w.TgBot.SendMessage(msg)
	if err != nil {
		return err
	}
	return nil
}

package worker

import (
	"fmt"
	"sync"

	"github.com/avraam311/delayed-notifier/internal/rabbitmq"
	"github.com/wb-go/wbf/zlog"
)

type Worker struct {
	RMQ          *rabbitmq.RabbitMq
	WorkersCount int
}

func New(RMQ *rabbitmq.RabbitMq, WorkersCount int) *Worker {
	return &Worker{
		RMQ:          RMQ,
		WorkersCount: WorkersCount,
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
	_, err := fmt.Println(msg)
	if err != nil {
		return err
	}
	return nil
}

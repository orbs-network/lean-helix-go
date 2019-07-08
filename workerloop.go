package leanhelix

import (
	"context"
)

type Worker struct {
}

func NewWorkerLoop() *Worker {
	return &Worker{}
}

func (w *Worker) Start(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return
		}
	}

}

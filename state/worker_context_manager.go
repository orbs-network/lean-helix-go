package state

import (
	"context"
	"sync"
)

type contextWithCancel struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type WorkerContextManager struct {
	hvToContext           map[HeightView]*contextWithCancel
	newestHvCanceledOlder *HeightView
	parentCtxWithCancel   *contextWithCancel
	mutex                 *sync.Mutex
}

func NewWorkerContextManager() *WorkerContextManager {
	return &WorkerContextManager{
		hvToContext:           make(map[HeightView]*contextWithCancel),
		newestHvCanceledOlder: nil,
		parentCtxWithCancel:   nil,
		mutex:                 &sync.Mutex{},
	}
}

func (w *WorkerContextManager) Init(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	w.parentCtxWithCancel = &contextWithCancel{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (w *WorkerContextManager) GetOrCreateContextFor(hv *HeightView) (context.Context, bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.newestHvCanceledOlder != nil && hv.OlderThan(w.newestHvCanceledOlder) {
		return nil, false
	}

	cc, ok := w.hvToContext[*hv]
	if !ok {
		ctx, cancel := context.WithCancel(w.parentCtxWithCancel.ctx)
		cc = &contextWithCancel{
			ctx:    ctx,
			cancel: cancel,
		}
		w.hvToContext[*hv] = cc
	}

	return cc.ctx, true
}

func (w *WorkerContextManager) CancelContextsOlderThan(hv *HeightView) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for chv, cc := range w.hvToContext {
		if chv.OlderThan(hv) {
			cc.cancel()
			delete(w.hvToContext, chv)
		}
	}

	if w.newestHvCanceledOlder == nil || w.newestHvCanceledOlder.OlderThan(hv) {
		w.newestHvCanceledOlder = hv
	}
}

func (w *WorkerContextManager) CancelAll() {
	w.parentCtxWithCancel.cancel()
}

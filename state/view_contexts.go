package state

import (
	"context"
	"fmt"
	"sync"
)

type contextWithCancel struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type ViewContexts struct {
	hvToContext           map[HeightView]*contextWithCancel
	newestHvCanceledOlder *HeightView
	parentCtxWithCancel   *contextWithCancel
	mutex                 *sync.Mutex
	shutdown              bool
}

func NewViewContexts() *ViewContexts {
	ctx, cancel := context.WithCancel(context.Background())
	return &ViewContexts{
		hvToContext:           make(map[HeightView]*contextWithCancel),
		newestHvCanceledOlder: nil,
		parentCtxWithCancel:   &contextWithCancel{
			ctx:    ctx,
			cancel: cancel,
		},
		mutex:                 &sync.Mutex{},
	}
}

func (w *ViewContexts) ActiveFor(hv *HeightView) (context.Context, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.shutdown {
		return nil, fmt.Errorf("shutting down")
	}

	if w.newestHvCanceledOlder != nil && hv.OlderThan(w.newestHvCanceledOlder) {
		return nil, fmt.Errorf("requested context for stale height/view %s", hv)
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

	return cc.ctx, nil
}

func (w *ViewContexts) CancelOlderThan(hv *HeightView) {
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

func (w *ViewContexts) Shutdown() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.parentCtxWithCancel.cancel()
	w.shutdown = true
}

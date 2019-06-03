package poc

import (
	"context"
	"fmt"
	"sync"
)

type LH struct {
	d                  *durations
	term               *Term
	updateStateChannel chan *Block
	messagesChannel    chan *Message
	// TODO: What to do with it when closing term?
	// TODO: Who creates this channel? LH or Term or NewLeanHelix()?
	committedChannel chan *Block
	currentHeight    int
}

type SPISender interface {
	sendMessage(m *Message)
}

func (lh *LH) sendMessage(m *Message) {
	lh.messagesChannel <- m
}

func NewLeanHelix(d *durations) *LH {
	Log("NewLeanHelix")
	return &LH{
		d:                  d,
		term:               nil,
		updateStateChannel: make(chan *Block, 0),
		messagesChannel:    make(chan *Message, 0),
		committedChannel:   make(chan *Block, 0),
		currentHeight:      1,
	}
}

func (lh *LH) StartLeanHelix(ctx context.Context, wg *sync.WaitGroup) {
	Log("lh.MainLoop StartLeanHelix starting *MAINLOOP* goroutine")

	id := ctx.Value("ID")
	newID := fmt.Sprintf("%s|MainLoop", id)
	// TODO Do something with cancel func?
	mainLoopCtx, _ := context.WithCancel(context.WithValue(ctx, "ID", newID))
	go lh.MainLoop(mainLoopCtx, wg)
}

func (lh *LH) MainLoop(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	Log("lh.MainLoop start ctx.ID=%s", ctx.Value("ID"))
	for {
		Log("lh.MainLoop IDLE")
		select {
		case <-ctx.Done():
			lh.shutdown()
			Log("lh.MainLoop end BYE")
			return
		case message := <-lh.messagesChannel:
			lh.filter(message)

		case block := <-lh.updateStateChannel:
			lh.resetTerm(ctx, wg, block)

		case block := <-lh.committedChannel:
			lh.resetTerm(ctx, wg, block)

		}
	}
}

func (lh *LH) shutdown() {
	Log("lh.shutdown")

	// TODO: CAREFUL: DO NOT CALL cancel(), it is already dead by the time it is called. Let term_ctx be cancelled automatically
	lh.term.cancel()
}

func (lh *LH) filter(message *Message) {
	Log("lh.filter msg=%s", message)
	lh.term.messagesChannel <- message
}

func (lh *LH) resetTerm(ctx context.Context, wg *sync.WaitGroup, block *Block) {
	Log("lh.resetTerm()")
	lh.cancelTerm()
	lh.startNewTerm(ctx, wg, block)
}

func (lh *LH) cancelTerm() {
	if lh.term != nil {
		Log("lh.cancelTerm() for H=%d", lh.term.height)
		lh.term.cancel()
	}
}

func (lh *LH) startNewTerm(parentCtx context.Context, wg *sync.WaitGroup, block *Block) {
	Log("lh.startNewTerm() start - block with H=%d", block.h)
	termMessagesChannel := make(chan *Message, 0)
	term := NewTerm(0, lh, block.h, termMessagesChannel, lh.committedChannel, lh.d.createBlock, lh.d.validateBlock)

	lh.term = term
	term.startTerm(parentCtx, wg)
	Log("lh.startNewTerm() end")
}

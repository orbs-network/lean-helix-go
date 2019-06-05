package poc

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type LH struct {
	d                  *Config
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

func NewLeanHelix(d *Config) *LH {
	Log("NewLeanHelix")
	return &LH{
		d:                  d,
		term:               nil,
		updateStateChannel: make(chan *Block, 0),
		messagesChannel:    make(chan *Message, d.MessageChannelBufLen),
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
		select {
		case <-ctx.Done():
			lh.shutdown()
			return

		case message := <-lh.messagesChannel:
			lh.onMessage(message)

		case block := <-lh.updateStateChannel:
			lh.onUpdateState(ctx, wg, block)

		case block := <-lh.committedChannel:
			lh.onCommit(ctx, wg, block)

		}
	}
}

func (lh *LH) onMessage(message *Message) {
	Log("lh.onMessage msg=%s", message)
	lh.filter(message)
}

func (lh *LH) onCommit(ctx context.Context, wg *sync.WaitGroup, block *Block) {
	lh.resetTerm(ctx, wg, block, true)
}

func (lh *LH) onUpdateState(ctx context.Context, wg *sync.WaitGroup, block *Block) {
	lh.resetTerm(ctx, wg, block, false)
}

func (lh *LH) shutdown() {
	Log("lh.shutdown MainLoop end BYE")
	lh.term.cancel()
}

func (lh *LH) filter(message *Message) {

	Log("FILTER sending msg=%s", message)
	lh.term.messagesChannel <- message
	Log("FILTER sent msg=%s", message)
}

func (lh *LH) resetTerm(ctx context.Context, wg *sync.WaitGroup, block *Block, fromCommit bool) {
	Log("lh.resetTerm()")
	lh.cancelTerm()
	Log("lh.resetTerm() starting term, will take 1 second")
	time.Sleep(1 * time.Second)
	lh.startNewTerm(ctx, wg, block, fromCommit)
	Log("lh.resetTerm() started term")
}

func (lh *LH) cancelTerm() {
	if lh.term != nil {
		Log("lh.cancelTerm() for H=%d", lh.term.height)
		lh.term.cancel()
	} else {
		Log("lh.cancelTerm() term is nil")
	}
}

func (lh *LH) startNewTerm(parentCtx context.Context, wg *sync.WaitGroup, block *Block, fromCommit bool) {
	Log("lh.startNewTerm() start - block with H=%d fromCommit=%t", block.h, fromCommit)
	termMessagesChannel := make(chan *Message, 0)
	term := NewTerm(0, lh, block.h, termMessagesChannel, lh.committedChannel, lh.d.CreateBlock, lh.d.ValidateBlock)

	lh.term = term
	term.startTerm(parentCtx, wg)
	Log("lh.startNewTerm() end")
}

func (lh *LH) electionNow() {
	if lh.term != nil {
		lh.term.ElectionNow()
	}
}

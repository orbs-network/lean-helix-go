package poc_listener_thr

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
	currentHeight      int
	electionChannel    chan interface{}
	electionTimer      *time.Timer
	cancelTermCtx      context.CancelFunc
}

type Term struct {
	instanceId    int
	height        int
	view          int
	cancelViewCtx context.CancelFunc
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

func (lh *LH) ListenerLoop(ctx context.Context, wg *sync.WaitGroup) {
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

		case <-lh.electionTimer.C:
			lh.onElection(ctx, wg, false)

		case <-lh.electionChannel: // for testing
			lh.onElection(ctx, wg, true)

		}
	}
}

func (lh *LH) MainLoop(ctx context.Context, wg *sync.WaitGroup) {
}

func (lh *LH) onMessage(message *Message) {
	Log("lh.onMessage msg=%s", message)
	lh.filter(message)
}

func (lh *LH) onElection(ctx context.Context, wg *sync.WaitGroup, manualTrigger bool) {
	Log("H=%d V=%d term.onElection manualTrigger=%t", lh.term.height, lh.term.view, manualTrigger)
	lh.term.setView(ctx, wg, lh.term.view+1)

}

func (lh *LH) onCommit(ctx context.Context, wg *sync.WaitGroup, block *Block) {
	lh.resetTerm(ctx, wg, block, true)
}

func (lh *LH) onUpdateState(ctx context.Context, wg *sync.WaitGroup, block *Block) {
	lh.resetTerm(ctx, wg, block, false)
}

func (lh *LH) shutdown() {
	Log("lh.shutdown MainLoop end BYE")
	lh.cancelTerm()
}

func (lh *LH) filter(message *Message) {
	Log("FILTER sending msg=%s", message)
	lh.messagesChannel <- message
	Log("FILTER sent msg=%s", message)
}

func (lh *LH) resetTerm(ctx context.Context, wg *sync.WaitGroup, block *Block, fromCommit bool) {
	Log("lh.resetTerm() starting term, will take 1 second")
	time.Sleep(1 * time.Second)
	lh.startNewTerm(ctx, wg, block, fromCommit)
	Log("lh.resetTerm() started term")
}

func (lh *LH) startNewTerm(parentCtx context.Context, wg *sync.WaitGroup, block *Block, fromCommit bool) {
	Log("lh.startNewTerm() start - block with H=%d fromCommit=%t", block.h, fromCommit)

	term := NewTerm(0, block.h)

	lh.term = term
	Log("lh.startNewTerm() end")
}

func (lh *LH) cancelTerm() {
	lh.term.cancel()
}

func NewTerm(instanceId int, height int) *Term {
	Log("NewTerm H=%d", height)
	newTerm := &Term{
		instanceId: instanceId,
		height:     height,
		view:       0,
	}
	return newTerm
}

func (term *Term) setView(ctx context.Context, wg *sync.WaitGroup, view int) {
	term.view = view
}

func (term *Term) cancel() {
	term.cancelView()
}

func (term *Term) cancelView() {
	term.cancelViewCtx()
}

package poc

import "context"

type LH struct {
	term               Term
	updateStateChannel chan *Block
	messagesChannel    chan *Message
	committedChannel   chan *Block
}

func NewLeanHelix() *LH {
	return &LH{}
}

func (lh *LH) MainLoop(ctx context.Context) {

	select {
	case <-ctx.Done():
		lh.shutdown()
		return
	case message := <-lh.messagesChannel:
		lh.filter(message)

	case block := <-lh.updateStateChannel:
		lh.resetTerm(ctx, block)

	case block := <-lh.committedChannel:
		lh.resetTerm(ctx, block)

	}

}

func (lh *LH) shutdown() {
	Log("lh.shutdown")
	lh.term.cancel()
}

func (lh *LH) filter(message *Message) {
	lh.term.messagesChannel <- message
}

func (lh *LH) resetTerm(ctx context.Context, block *Block) {
	lh.cancelTerm()
	lh.startNewTerm(ctx, block)

}

func (lh *LH) cancelTerm() {
	lh.term.cancel()
}

func (lh *LH) startNewTerm(parentCtx context.Context, block *Block) {
	termMessagesChannel := make(chan *Message)
	term := NewTerm(block.h, termMessagesChannel)
	ctx, cancel := context.WithCancel(parentCtx)
	term.cancel = cancel
	go term.termloop(ctx)
}

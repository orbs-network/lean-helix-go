package poc

import "context"

type LH struct {
	term               *Term
	updateStateChannel chan *Block
	messagesChannel    chan *Message
	// TODO: What to do with it when closing term?
	// TODO: Who creates this channel? LH or Term or NewLeanHelix()?
	committedChannel chan *Block
}

func NewLeanHelix() *LH {
	Log("NewLeanHelix")
	return &LH{
		term:               nil,
		updateStateChannel: make(chan *Block, 0),
		messagesChannel:    make(chan *Message, 0),
		committedChannel:   make(chan *Block, 0),
	}
}

func (lh *LH) MainLoop(ctx context.Context) {
	Log("lh.MainLoop start")
	for {
		select {
		case <-ctx.Done():
			lh.shutdown()
			Log("lh.MainLoop end")
			return
		case message := <-lh.messagesChannel:
			lh.filter(message)

		case block := <-lh.updateStateChannel:
			lh.resetTerm(ctx, block)

		case block := <-lh.committedChannel:
			lh.resetTerm(ctx, block)

		}
	}
}

func (lh *LH) shutdown() {
	Log("lh.shutdown")

	// TODO: CAREFUL: DO NOT CALL cancel(), it is already dead by the time it is called. Let term_ctx be cancelled automatically
	// lh.term.cancel()
}

func (lh *LH) filter(message *Message) {
	Log("lh.filter")
	lh.term.messagesChannel <- message
}

func (lh *LH) resetTerm(ctx context.Context, block *Block) {
	Log("lh.resetTerm")
	lh.cancelTerm()
	lh.startNewTerm(ctx, block)

}

func (lh *LH) cancelTerm() {
	Log("lh.cancelTerm")
	if lh.term != nil && lh.term.cancel != nil {
		lh.term.cancel()
	}
}

func (lh *LH) startNewTerm(parentCtx context.Context, block *Block) {
	Log("lh.startNewTerm")
	termMessagesChannel := make(chan *Message, 0)
	term := NewTerm(0, block.h, termMessagesChannel, lh.committedChannel)
	ctx, cancel := context.WithCancel(parentCtx)
	term.cancel = cancel
	go term.TermLoop(ctx)
}

package poc_2thr

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

type OnCommitCallback func(*Block)

type Term struct {
	instanceId      int
	sender          SPISender
	cancel          context.CancelFunc
	messagesChannel chan *Message
	height          int
	view            int
	electionChannel chan interface {
	}
	electionTimer *time.Timer
	// TODO: What to do with it when closing term?
	// TODO: Who creates this channel?
	asyncOpChannel chan *Block
	// TODO: What to do with it when closing term?
	// TODO: Who creates this channel?
	validateBlockChannel  chan bool
	onCommitCallback      OnCommitCallback
	createBlockDuration   time.Duration
	validateBlockDuration time.Duration
	cancelCreateBlock     context.CancelFunc
}

type ElectionTrigger struct {
	notificationChannel chan interface{}
}

func NewTerm(instanceId int, sender SPISender, height int, ch chan *Message, onCommitCallback OnCommitCallback, timeToCreateBlock time.Duration, timeToValidateBlock time.Duration) *Term {
	Log("NewTerm H=%d", height)
	newTerm := &Term{
		instanceId:            instanceId,
		sender:                sender,
		cancel:                nil,
		messagesChannel:       ch,
		height:                height,
		view:                  0,
		electionChannel:       nil,
		electionTimer:         nil,
		asyncOpChannel:        nil,
		validateBlockChannel:  nil,
		onCommitCallback:      onCommitCallback,
		createBlockDuration:   timeToCreateBlock,
		validateBlockDuration: timeToValidateBlock,
	}
	return newTerm
}

func (term *Term) isLeader() bool {
	return true // no notion of leader/non-leader yet
}

// This assumes this is always leader
func (term *Term) TermLoop(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	Log("H=%d term.TermLoop start. Ctx.ID=%s", term.height, ctx.Value("ID"))
	term.setView(ctx, wg, 0)

	// Sync version
	//block := term.CreateBlock()
	//term.sendMessage(NewPPM(block))

	// Async version

	for {
		Log("H=%d V=%d term.TermLoop IDLE", term.height, term.view)
		select {
		case <-ctx.Done():
			Log("H=%d V=%d term.TermLoop ctx.Done BYE", term.height, term.view)
			return

		case message := <-term.messagesChannel:
			term.onMessage(message)

		// TODO: What happens if this pops just after closing the term?
		case <-term.electionTimer.C:
			term.onElection(ctx, wg, false)

		case <-term.electionChannel: // for testing
			term.onElection(ctx, wg, true)

		case block := <-term.asyncOpChannel:
			term.onEndedCreateBlock(block)

		case validationResult := <-term.validateBlockChannel:
			term.onEndedValidateBlock(validationResult)

		}
	}
}

func (term *Term) onMessage(message *Message) {
	Log("H=%d V=%d  >>> MSG: %s", term.height, term.view, message)

	switch message.msgType {
	case COMMIT:
		term.onCommit(message.block)
	}
}

func (term *Term) onElection(ctx context.Context, wg *sync.WaitGroup, manualTrigger bool) {
	Log("H=%d V=%d term.onElection manualTrigger=%t", term.height, term.view, manualTrigger)
	term.setView(ctx, wg, term.view+1)

}

func (term *Term) onCommit(blockToCommit *Block) {
	Log("H=%d V=%d term.onCommit sending", term.height, term.view)
	term.onCommitCallback(blockToCommit)
	Log("H=%d V=%d term.onCommit sent", term.height, term.view)
}

func calcElectionTimeout(view int) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(2.0, float64(view))))
	return timeoutMultiplier * time.Millisecond * 100
}

func (term *Term) setView(ctx context.Context, wg *sync.WaitGroup, view int) {
	//Log("H=%d term.setView(V=%d)", term.height, view)
	if term.cancelCreateBlock != nil {
		term.cancelCreateBlock()
	}
	term.view = view
	if term.electionTimer != nil {
		Log("H=%d term.setView(V=%d) electionTimer.stop", term.height, view)
		term.electionTimer.Stop()
	}
	timeout := calcElectionTimeout(view)
	term.electionTimer = time.NewTimer(timeout)
	term.electionChannel = make(chan interface{})
	Log("H=%d term.setView(V=%d) new electionTimer timeout=%s", term.height, view, timeout)

	term.maybeCreateBlock(ctx, wg)
}

func (term *Term) maybeCreateBlock(ctx context.Context, wg *sync.WaitGroup) {
	if !term.isLeader() {
		return
	}
	ctxId := ctx.Value("ID")
	newCtxID := fmt.Sprintf("%s|V=%d|CreateBlock", ctxId, term.view)
	// TODO Do something with cancel func?
	createBlockCtx, cancelCreateBlock := context.WithCancel(context.WithValue(ctx, "ID", newCtxID))
	term.asyncOpChannel = make(chan *Block)
	term.cancelCreateBlock = cancelCreateBlock
	go CreateBlock(createBlockCtx, wg, term.asyncOpChannel, term.height, term.view, term.createBlockDuration)

}

func (term *Term) startTerm(parentCtx context.Context, wg *sync.WaitGroup) {

	Log("H=%d term.startTerm() starting *TERMLOOP* goroutine", term.height)
	id := parentCtx.Value("ID")
	newID := fmt.Sprintf("%s|TermLoop_H=%d", id, term.height)
	// TODO Do something with cancel func?
	mainLoopCtx, cancel := context.WithCancel(context.WithValue(parentCtx, "ID", newID))
	term.cancel = cancel

	go term.TermLoop(mainLoopCtx, wg)
}

func (term *Term) onEndedCreateBlock(block *Block) {
	Log("H=%d V=%d term.onEndedCreateBlock %s", term.height, block, term.view)
}

func (term *Term) onEndedValidateBlock(b bool) {
	Log("H=%d V=%d term.onEndedValidateBlock", term.height, term.view)
}

func (term *Term) sendMessage(m *Message) {
	term.sender.sendMessage(m)
}

func (term *Term) ElectionNow() {
	Log("++++++ H=%d V=%d term.ElectionNow() closing channel", term.height, term.view)
	close(term.electionChannel)
	Log("++++++ H=%d V=%d term.ElectionNow() closed channel", term.height, term.view)
}

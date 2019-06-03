package poc

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

type Term struct {
	id              int
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
	createBlockChannel chan *Block
	// TODO: What to do with it when closing term?
	// TODO: Who creates this channel?
	validateBlockChannel  chan bool
	committedChannel      chan *Block
	createBlockDuration   time.Duration
	validateBlockDuration time.Duration
}

type ElectionTrigger struct {
	notificationChannel chan interface{}
}

func NewTerm(id int, sender SPISender, height int, ch chan *Message, commitCh chan *Block, timeToCreateBlock time.Duration, timeToValidateBlock time.Duration) *Term {
	Log("NewTerm H=%d", height)
	newTerm := &Term{
		id:                    id,
		sender:                sender,
		cancel:                nil,
		messagesChannel:       ch,
		height:                height,
		view:                  0,
		electionChannel:       make(chan interface{}),
		electionTimer:         nil,
		createBlockChannel:    make(chan *Block),
		validateBlockChannel:  make(chan bool),
		committedChannel:      commitCh,
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
	term.setView(0)

	// Sync version
	//block := term.CreateBlock()
	//term.sendMessage(NewPPM(block))

	// Async version
	id := ctx.Value("ID")
	newID := fmt.Sprintf("%s|CreateBlock", id)
	// TODO Do something with cancel func?
	createBlockCtx, cancelCreateBlock := context.WithCancel(context.WithValue(ctx, "ID", newID))
	go CreateBlock(createBlockCtx, wg, term.createBlockChannel, term.height, term.createBlockDuration)

	for {
		Log("H=%d term.TermLoop IDLE", term.height)
		select {
		case <-ctx.Done():
			Log("H=%d term.TermLoop ctx.Done BYE", term.height)
			cancelCreateBlock()
			return

		case message := <-term.messagesChannel:
			term.onMessage(message)

		// TODO: What happens if this pops just after closing the term?
		case <-term.electionTimer.C:
			term.onElection()

		case block := <-term.createBlockChannel:
			term.onEndedCreateBlock(block)

		case validationResult := <-term.validateBlockChannel:
			term.onEndedValidateBlock(validationResult)

		}
	}
}

func (term *Term) onMessage(message *Message) {
	Log("term.onMessage %s", message)

	switch message.msgType {
	case COMMIT:
		term.onCommit(message.block)
	}
}

func (term *Term) onElection() {
	Log("H=%d term.onElection V=%d", term.height, term.view)
	term.setView(term.view + 1)

}

func (term *Term) onCommit(blockToCommit *Block) {
	Log("H=%d term.onCommit", term.height)
	term.committedChannel <- blockToCommit
}

func calcElectionTimeout(view int) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(2.0, float64(view))))
	return timeoutMultiplier * time.Millisecond * 100
}

func (term *Term) setView(view int) {
	Log("H=%d term.setView(V=%d)", term.height, view)
	term.view = view
	if term.electionTimer != nil {
		Log("H=%d term.setView(V=%d) electionTimer.stop", term.height, view)
		term.electionTimer.Stop()
	}
	timeout := calcElectionTimeout(view)
	term.electionTimer = time.NewTimer(timeout)
	Log("H=%d term.setView(V=%d) new electionTimer timeout=%s", term.height, view, timeout)
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
	Log("H=%d term.onEndedCreateBlock %s", term.height, block)
}

func (term *Term) onEndedValidateBlock(b bool) {
	Log("H=%d term.onEndedValidateBlock", term.height)
}

func (term *Term) sendMessage(m *Message) {
	term.sender.sendMessage(m)
}

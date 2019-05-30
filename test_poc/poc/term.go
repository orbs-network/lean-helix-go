package poc

import (
	"context"
	"math"
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
		height:                height,
		view:                  0,
		messagesChannel:       ch,
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
func (term *Term) TermLoop(ctx context.Context) {
	Log("H=%d term.TermLoop start", term.height)
	term.setView(0)

	// Sync version
	//block := term.CreateBlock()
	//term.sendMessage(NewPPM(block))

	// Async version
	go term.CreateBlock(ctx, term.createBlockChannel)

	for {
		Log("H=%d term.TermLoop IDLE", term.height)
		select {
		case <-ctx.Done():
			Log("H=%d term.TermLoop ctx.Done BYE", term.height)
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

func (term *Term) startTerm(ctx context.Context) {

	Log("H=%d term.startTerm() starting *TERMLOOP* goroutine", term.height)
	go term.TermLoop(ctx)
}

func (term *Term) onEndedCreateBlock(block *Block) {
	Log("H=%d term.onEndedCreateBlock %s", term.height, block)
}

func (term *Term) onEndedValidateBlock(b bool) {
	Log("H=%d term.onEndedCreateBlock", term.height)
}

// Let's assume this can't be interrupted during execution
// (in reality it can, but this assumes worst case behavior of external service)
func (term *Term) CreateBlock(ctx context.Context, responseChannel chan *Block) {
	Log("H=%d CREATE_BLOCK term.CreateBlock() duration=%s start", term.height, term.createBlockDuration)
	time.Sleep(term.createBlockDuration)
	Log("H=%d CREATE_BLOCK term.CreateBlock() duration=%s end, pushing to response channel", term.height, term.createBlockDuration)
	responseChannel <- NewBlock(term.height)
	Log("H=%d CREATE_BLOCK term.CreateBlock() duration=%s end, pushed to response channel", term.height, term.createBlockDuration)
}

func (term *Term) sendMessage(m *Message) {
	term.sender.sendMessage(m)
}

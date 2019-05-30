package poc

import (
	"context"
	"math"
	"time"
)

type Term struct {
	id              int
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
	validateBlockChannel chan bool
	committedChannel     chan *Block
}

type ElectionTrigger struct {
	notificationChannel chan interface{}
}

func NewTerm(id int, height int, ch chan *Message, commitCh chan *Block) *Term {
	Log("NewTerm H=%d", height)
	newTerm := &Term{
		id:               id,
		height:           height,
		view:             0,
		messagesChannel:  ch,
		committedChannel: commitCh,
	}

	newTerm.setView(0)
	return newTerm
}

func (term *Term) isLeader() bool {
	return true // no notion of leader/non-leader yet
}

func (term *Term) TermLoop(ctx context.Context) {
	Log("term.TermLoop")
	for {
		select {
		case <-ctx.Done():
			Log("term.TermLoop ctx.Done")
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
	Log("term.onElection")
	term.setView(term.view + 1)

}

func (term *Term) onCommit(blockToCommit *Block) {
	Log("term.onCommit")
	term.committedChannel <- blockToCommit
}

func calcElectionTimeout(view int) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(2.0, float64(view))))
	return timeoutMultiplier * time.Millisecond * 100
}

func (term *Term) setView(view int) {
	term.view = view
	if term.electionTimer != nil {
		Log("term.setView %d electionTimer.stop", view)
		term.electionTimer.Stop()
	}
	timeout := calcElectionTimeout(view)
	term.electionTimer = time.NewTimer(timeout)
	Log("term.setView %d new electionTimer timeout=%s", view, timeout)
}

func (term *Term) onEndedCreateBlock(block *Block) {
	Log("term.onEndedCreateBlock %s", block)
}

func (term *Term) onEndedValidateBlock(b bool) {
	Log("term.onEndedCreateBlock")
}

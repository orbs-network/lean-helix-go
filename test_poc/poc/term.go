package poc

import (
	"context"
	"math"
	"time"
)

type Term struct {
	cancel               context.CancelFunc
	messagesChannel      chan *Message
	height               int
	view                 int
	createBlockChannel   chan *Block
	electionChannel      chan interface{}
	electionTimer        *time.Timer
	validateBlockChannel chan bool
}

type ElectionTrigger struct {
	notificationChannel chan interface{}
}

func NewTerm(height int, ch chan *Message) *Term {
	newTerm := &Term{
		height:          height,
		view:            0,
		messagesChannel: ch,
	}

	newTerm.setView(0)
	return newTerm
}

func (term *Term) termloop(ctx context.Context) {
	select {
	case <-ctx.Done():
		return

	case message := <-term.messagesChannel:
		term.onMessage(message)

	case <-term.electionTimer.C:
		term.onElection()

	case block := <-term.createBlockChannel:
		term.onEndedCreateBlock(block)

	case validationResult := <-term.validateBlockChannel:
		term.onEndedValidateBlock(validationResult)

	}

}

func (term *Term) onMessage(message *Message) {

}

func (term *Term) onElection() {
	term.setView(term.view + 1)

}

func calcElectionTimeout(view int) time.Duration {
	timeoutMultiplier := time.Duration(int64(math.Pow(2.0, float64(view))))
	return timeoutMultiplier * time.Millisecond * 100
}

func (term *Term) setView(view int) {
	term.view = view
	term.electionTimer.Stop()
	term.electionTimer = time.NewTimer(calcElectionTimeout(view))
}

func (term *Term) onEndedCreateBlock(block *Block) {

}

func (term *Term) onEndedValidateBlock(b bool) {

}

package poc

import (
	"context"
	"sync"
	"time"
)

func runOrbs(ctx context.Context, wg *sync.WaitGroup, d *durations) {

	wg.Add(1)
	defer wg.Done()
	Log("TTTT TEST runOrbs() start")
	lh := NewLeanHelix(d)
	lh.StartLeanHelix(ctx, wg)

	time.Sleep(150 * time.Millisecond)
	Log("TTTT TEST runOrbs() UpdateState 1")
	updateFromNodeSync(lh, 1)

	time.Sleep(150 * time.Millisecond)
	Log("TTTT TEST runOrbs() UpdateState 2")
	updateFromNodeSync(lh, 2)

	time.Sleep(150 * time.Millisecond)
	Log("TTTT TEST runOrbs() COMMIT 3")
	go sendMessage(lh, NewCM(NewBlock(3)))
	Log("TTTT TEST runOrbs() sent commit")
	time.Sleep(5 * time.Millisecond)

	go electionNow(lh)
	Log("TTTT TEST runOrbs() sent election")

}

func electionNow(lh *LH) {
	lh.electionNow()
}

func updateFromNodeSync(lh *LH, height int) {
	b := NewBlock(height)
	timer := time.AfterFunc(1000*time.Millisecond, myPanic)
	Log("ORBS updateFromNodeSync() H=%d sending to updateChannel", height)
	lh.updateStateChannel <- b
	Log("ORBS updateFromNodeSync() H=%d sent to updateChannel", height)
	timer.Stop()
}

// Let's assume this can't be interrupted during execution
// (in reality it can, but this assumes worst case behavior of external service)
func CreateBlock(ctx context.Context, wg *sync.WaitGroup, responseChannel chan *Block, height int, view int, createBlockDuration time.Duration) {
	wg.Add(1)
	defer wg.Done()

	Log("H=%d V=%d CREATE_BLOCK term.CreateBlock() duration=%s start ctx.ID=%s", height, view, createBlockDuration, ctx.Value("ID"))
	time.Sleep(createBlockDuration)
	Log("H=%d V=%d CREATE_BLOCK term.CreateBlock() duration=%s end, pushing to response channel", height, view, createBlockDuration)
	select {
	case <-ctx.Done():
		Log("H=%d V=%d CREATE_BLOCK CANCELLED ctx.ID=%s", height, view, ctx.Value("ID"))
	case responseChannel <- NewBlock(height):
		Log("H=%d V=%d CREATE_BLOCK term.CreateBlock() duration=%s end, pushed to response channel", height, view, createBlockDuration)
	default:
	}

}

func sendMessage(lh *LH, m *Message) {
	lh.messagesChannel <- m
}

func myPanic() {
	panic("WAITED FOR TOO LONG!")
}

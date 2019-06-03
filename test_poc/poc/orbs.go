package poc

import (
	"context"
	"sync"
	"time"
)

func runOrbs(ctx context.Context, wg *sync.WaitGroup, d *durations) {

	wg.Add(1)
	defer wg.Done()
	Log("runOrbs() start")
	lh := NewLeanHelix(d)
	lh.StartLeanHelix(ctx, wg)

	time.Sleep(200 * time.Millisecond)
	updateFromNodeSync(lh, 1)

	time.Sleep(200 * time.Millisecond)
	updateFromNodeSync(lh, 2)

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
func CreateBlock(ctx context.Context, wg *sync.WaitGroup, responseChannel chan *Block, height int, createBlockDuration time.Duration) {
	wg.Add(1)
	defer wg.Done()

	Log("H=%d CREATE_BLOCK term.CreateBlock() duration=%s start ctx.ID=%s", height, createBlockDuration, ctx.Value("ID"))
	time.Sleep(createBlockDuration)
	Log("H=%d CREATE_BLOCK term.CreateBlock() duration=%s end, pushing to response channel", height, createBlockDuration)
	select {
	case <-ctx.Done():
		Log("H=%d CREATE_BLOCK CANCELLED ctx.ID=%s", height, ctx.Value("ID"))
	case responseChannel <- NewBlock(height):
		Log("H=%d CREATE_BLOCK term.CreateBlock() duration=%s end, pushed to response channel", height, createBlockDuration)
	default:
	}

}

func sendMessage(lh *LH, m *Message) {
	lh.messagesChannel <- m
}

func myPanic() {
	panic("WAITED FOR TOO LONG!")
}

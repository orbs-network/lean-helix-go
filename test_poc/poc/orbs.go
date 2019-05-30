package poc

import (
	"context"
	"time"
)

func runOrbs(ctx context.Context, d *durations) {

	Log("runOrbs() start")
	lh := NewLeanHelix(d)
	lh.StartLeanHelix(ctx)

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

func sendMessage(lh *LH, m *Message) {
	lh.messagesChannel <- m
}

func myPanic() {
	panic("WAITED FOR TOO LONG!")
}

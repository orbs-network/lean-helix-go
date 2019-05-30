package poc

import (
	"context"
	"time"
)

func runOrbs(ctx context.Context) {

	Log("runOrbs() start")
	lh := NewLeanHelix()
	go lh.MainLoop(ctx)

	time.Sleep(200 * time.Millisecond)
	updateFromNodeSync(lh, 1)

	time.Sleep(200 * time.Millisecond)
	updateFromNodeSync(lh, 2)

}

func updateFromNodeSync(lh *LH, height int) {
	b := NewBlock(height)
	Log("ORBS updateFromNodeSync() H=%d send to updateChannel", height)
	lh.updateStateChannel <- b
}

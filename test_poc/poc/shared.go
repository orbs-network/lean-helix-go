package poc

import (
	"context"
	"fmt"
	"time"
)

type MessageType int

const PREPREPARE MessageType = 1
const PREPARE MessageType = 2
const COMMIT MessageType = 3
const VIEW_CHANGE MessageType = 4
const NEW_VIEW MessageType = 5

type Message struct {
	msgType MessageType
	block   *Block
}

type Block struct {
	h int
}

func Run() {
	timeToCancel := 2 * time.Second
	Log("Run() start timeToCancel=%s", timeToCancel)
	ctx, cancel := context.WithCancel(context.Background())
	go runOrbs(ctx)
	time.Sleep(timeToCancel)
	Log("Run() CANCELLING")
	cancel()
	time.Sleep(500 * time.Millisecond)
	Log("Run() end")
}

func NewBlock(h int) *Block {
	return &Block{
		h: h,
	}
}

func Log(format string, a ...interface{}) {
	fmt.Printf(time.Now().Format("03:04:05.000000")+" "+format+"\n", a...)
}

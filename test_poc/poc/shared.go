package poc

import (
	"context"
	"fmt"
	"runtime"
	"sync"
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

type durations struct {
	cancelTestAfter     time.Duration
	waitAfterCancelTest time.Duration
	createBlock         time.Duration
	validateBlock       time.Duration
}

// Using WaitGroups
// See https://nathanleclaire.com/blog/2014/02/15/how-to-wait-for-all-goroutines-to-finish-executing-before-continuing/
func Run(d *durations) {
	timeToCancel := d.cancelTestAfter
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), "ID", "ROOT"))
	Log("Run() start timeToCancel=%s starting *ORBS* goroutine", timeToCancel)
	var wg sync.WaitGroup
	go runOrbs(ctx, &wg, d)
	time.Sleep(timeToCancel)
	Log("Run() CANCELLING TEST ON TIMEOUT")
	cancel()
	time.Sleep(d.waitAfterCancelTest)
	numGoroutineAfter := runtime.NumGoroutine()
	Log("Goroutines=%d WAIT FOR WG", numGoroutineAfter)
	wg.Wait()
	Log("********** Run() end. Goroutines=%d", numGoroutineAfter)
}

func NewBlock(h int) *Block {
	return &Block{
		h: h,
	}
}

func Log(format string, a ...interface{}) {
	fmt.Printf(time.Now().Format("03:04:05.000")+" "+format+"\n", a...)
}

func NewPPM(block *Block) *Message {
	return &Message{
		msgType: PREPREPARE,
		block:   block,
	}
}

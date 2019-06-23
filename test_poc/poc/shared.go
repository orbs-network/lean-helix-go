package poc

import (
	"context"
	"fmt"
	"time"
)

type Message struct {
	msgType int
}

type Block struct {
	h int
}

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	lh := NewLeanHelix()
	go lh.MainLoop(ctx)
	time.Sleep(5 * time.Second)
}

func Log(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}

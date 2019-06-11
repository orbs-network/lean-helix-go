package main

import (
	"github.com/orbs-network/lean-helix-go/test_poc/poc_2thr"
	"time"
)

func main() {

	config := &poc_2thr.Config{
		CancelTestAfter:     1000 * time.Millisecond,
		WaitAfterCancelTest: 500 * time.Millisecond,
		CreateBlock:         500 * time.Millisecond,
		ValidateBlock:       500 * time.Millisecond,
	}

	poc_2thr.Run(config)
}

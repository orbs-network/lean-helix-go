package poc

import (
	"testing"
	"time"
)

// See README.md for POC docs

func TestMainFlow(t *testing.T) {

	d := &Config{
		CancelTestAfter:      2000 * time.Millisecond,
		WaitAfterCancelTest:  500 * time.Millisecond,
		CreateBlock:          500 * time.Millisecond,
		ValidateBlock:        500 * time.Millisecond,
		MessageChannelBufLen: 10,
	}

	Run(d)
}

func TestGarbageCollectedChan(t *testing.T) {

}

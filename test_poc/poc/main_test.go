package poc

import (
	"testing"
	"time"
)

// See README.md for POC docs

func TestMainFlow(t *testing.T) {

	d := &durations{
		cancelTestAfter:     3000 * time.Millisecond,
		waitAfterCancelTest: 500 * time.Millisecond,
		createBlock:         1000 * time.Millisecond,
		validateBlock:       1000 * time.Millisecond,
	}

	Run(d)
}

func TestGarbageCollectedChan(t *testing.T) {

}

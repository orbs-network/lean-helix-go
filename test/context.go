// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"sync"
	"testing"
	"time"
)

func WithContext(f func(ctx context.Context)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f(ctx)
}

func WithContextWithTimeout(t *testing.T, d time.Duration, f func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	f(ctx)
	if ctx.Err() != nil {
		panic("WithContextWithTimeout() timed out")
	}
}

func FailIfNotDoneByTimeout(t *testing.T, waitGroup *sync.WaitGroup, timeout time.Duration, format string, args ...interface{}) {
	timeoutCtx, _ := context.WithTimeout(context.Background(), timeout)

	condDone := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(condDone)
	}()

	select {
	case <-condDone: // wait group finished waiting
	case <-timeoutCtx.Done():
		t.Fatalf(format, args...)
	}
}

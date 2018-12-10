package leaderelection

import (
	"context"
	"github.com/orbs-network/orbs-network-go/test"
	"testing"
)

func TestLeaderCircularOrdering(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := NewHarness()

		h.TriggerElection()
		h.waitForLeader(0, ctx)

		h.TriggerElection()
		h.waitForLeader(1, ctx)

		h.TriggerElection()
		h.waitForLeader(2, ctx)

		h.TriggerElection()
		h.waitForLeader(3, ctx)

		h.TriggerElection()
		h.waitForLeader(0, ctx)
	})
}

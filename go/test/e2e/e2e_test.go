package e2e

import (
	"fmt"
	"github.com/orbs-network/orbs-spec/types/go/services"
	"github.com/stretchr/testify/mock"
	"testing"
)

// Adapted from PBFT.spec.ts

const NODE_COUNT = 1

func TestSendPreprepareOnlyIfLeader(t *testing.T) {

	h := NewHarness(NODE_COUNT)

	h.service.network.When("sendToMembers", &services.CommitBlockInput{expectedBlockPair}).Return(nil, nil).Times(1)
	h.service.network.When("SendBenchmarkConsensusCommitted", mock.AnyIf(fmt.Sprintf("LastCommittedBlockHeight equals %d, recipient equals %s and sender equals %s", expectedLastCommitted, expectedRecipient, expectedSender), lastCommittedReplyMatcher)).Times(1)
	h.service.Start()
	defer h.service.Stop()

}

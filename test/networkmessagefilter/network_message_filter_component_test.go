package networkmessagefilter

import (
	"github.com/orbs-network/lean-helix-go/primitives"
	"testing"
)

const NODE_COUNT = 4

func TestSetBlockHeightAndReceiveGossipMessages(t *testing.T) {
	messagesBlockHeight := primitives.BlockHeight(3)
	nodesBlockHeight := primitives.BlockHeight(3)
	messagesView := primitives.View(0)
	senderNodeIndex := 0

	h := NewHarness(NODE_COUNT, senderNodeIndex, messagesBlockHeight, nodesBlockHeight, messagesView, nil)
	h.ExpectEachMessageToBeReceivedXTimes(1, []int{senderNodeIndex})
	h.ExpectXMessagesToBeSent(5)

	if err := h.SendAllMessages(); err != nil {
		t.Error(err)
	}
	ok, err := h.Verify()
	if !ok {
		t.Error(err)
	}
}

func TestIgnoreMessagesNotFromCurrentBlockHeight(t *testing.T) {
	messagesBlockHeight := primitives.BlockHeight(2)
	filterBlockHeight := primitives.BlockHeight(3)
	senderNodeIndex := 0

	h := NewHarness(NODE_COUNT, senderNodeIndex, messagesBlockHeight, filterBlockHeight, primitives.View(0), nil)
	h.ExpectEachMessageToBeReceivedXTimes(0, []int{senderNodeIndex})
	h.ExpectXMessagesToBeSent(5)

	if err := h.SendAllMessages(); err != nil {
		t.Error(err)
	}
	ok, err := h.Verify()
	if !ok {
		t.Error(err)
	}
}

func TestIgnoreMessagesFromMyself(t *testing.T) {
	messagesBlockHeight := primitives.BlockHeight(3)
	filterBlockHeight := primitives.BlockHeight(3)
	senderNodeIndex := 0

	// All excluded - the sender because it's not supposed to receive
	excludedReceivers := make([]int, NODE_COUNT-1)
	for i := 0; i < NODE_COUNT-1; i++ {
		excludedReceivers[i] = i + 1
	}

	h := NewHarness(NODE_COUNT, senderNodeIndex, messagesBlockHeight, filterBlockHeight, primitives.View(0), nil)
	h.ExpectEachMessageToBeReceivedXTimes(0, excludedReceivers)
	h.ExpectEachMessageToBeReceivedXTimes(1, []int{senderNodeIndex})
	h.ExpectXMessagesToBeSent(5)

	if err := h.SendAllMessages(); err != nil {
		t.Error(err)
	}
	ok, err := h.Verify()
	if !ok {
		t.Error(err)
	}
}

func TestIgnoreMessagesFromNodesNotPartOfTheNetwork(t *testing.T) {

	t.Skip()
	messagesBlockHeight := primitives.BlockHeight(3)
	filterBlockHeight := primitives.BlockHeight(3)
	nonMemberSenderNodeIndex := 2
	nonMemberNodeIndices := []int{nonMemberSenderNodeIndex}

	h := NewHarness(NODE_COUNT, nonMemberSenderNodeIndex, messagesBlockHeight, filterBlockHeight, primitives.View(0), nonMemberNodeIndices)
	h.ExpectEachMessageToBeReceivedXTimes(0, nil)
	h.ExpectXMessagesToBeSent(5)

	if err := h.SendAllMessages(); err != nil {
		t.Error(err)
	}
	ok, err := h.Verify()
	if !ok {
		t.Error(err)
	}
}

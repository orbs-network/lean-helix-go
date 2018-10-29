package networkmessagefilter

import (
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestSetBlockHeightAndReceiveGossipMessages(t *testing.T) {
	messagesBlockHeight := primitives.BlockHeight(3)
	blockHeight := primitives.BlockHeight(3)

	h := NewHarness(blockHeight)
	h.GenerateMessages(messagesBlockHeight, h.senderNode.KeyManager)
	h.ExpectEachMessageToBeReceivedXTimes(1)

	verify(h, t)
}

func TestIgnoreMessagesNotFromCurrentBlockHeight(t *testing.T) {
	messagesBlockHeight := primitives.BlockHeight(2)
	blockHeight := primitives.BlockHeight(3)

	h := NewHarness(blockHeight)
	h.GenerateMessages(messagesBlockHeight, h.senderNode.KeyManager)
	h.ExpectEachMessageToBeReceivedXTimes(0)

	verify(h, t)
}

func TestIgnoreMessagesFromMyself(t *testing.T) {
	messagesBlockHeight := primitives.BlockHeight(3)
	blockHeight := primitives.BlockHeight(3)

	h := NewHarness(blockHeight)
	h.GenerateMessages(messagesBlockHeight, h.receiverNode.KeyManager)
	h.ExpectEachMessageToBeReceivedXTimes(0)

	verify(h, t)
}

func TestIgnoreMessagesFromNodesNotPartOfTheNetwork(t *testing.T) {
	messagesBlockHeight := primitives.BlockHeight(3)
	blockHeight := primitives.BlockHeight(3)

	dummyPublicKey := primitives.Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	dummyKeyManager := builders.NewMockKeyManager(dummyPublicKey)
	h := NewHarness(blockHeight)
	h.GenerateMessages(messagesBlockHeight, dummyKeyManager)
	h.ExpectEachMessageToBeReceivedXTimes(0)

	verify(h, t)
}

func verify(h *harness, t *testing.T) {
	if err := h.SendAllMessages(); err != nil {
		t.Error(err)
	}
	ok, err := h.Verify()
	if !ok {
		t.Error(err)
	}
}

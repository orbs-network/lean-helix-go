package networkmessagefilter

import (
	"context"
	"fmt"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

const MINIMUM_NODE_COUNT = 4

type harness struct {
	ctx             context.Context
	net             *builders.TestNetwork
	senderNodeIndex int
	senderGossip    *gossip.Gossip
	filter          *leanhelix.NetworkMessageFilter
	block           leanhelix.Block
	messages        []leanhelix.ConsensusMessage
	allPublicKeys   []primitives.Ed25519PublicKey
}

func NewHarness(
	nodeCount int,
	senderNodeIndex int,
	messagesBlockHeight primitives.BlockHeight,
	nodesBlockHeight primitives.BlockHeight,
	messagesView primitives.View,
	nonMemberNodeIndices []int) *harness {

	if nodeCount < MINIMUM_NODE_COUNT {
		panic(fmt.Sprintf("minimum node count is %d", MINIMUM_NODE_COUNT))
	}

	ctx := context.Background()

	net := builders.NewSimpleTestNetwork(nodeCount, nodesBlockHeight, nil, nonMemberNodeIndices) // Node 0 is leader
	senderNode := net.Nodes[senderNodeIndex]

	senderMessageFactory := leanhelix.NewMessageFactory(senderNode.KeyManager)
	senderGossip := net.GetNodeGossip(senderNode.KeyManager.MyPublicKey())
	block := builders.CreateBlock(builders.GenesisBlock)
	messages := make([]leanhelix.ConsensusMessage, 5)
	messages[0] = senderMessageFactory.CreatePreprepareMessage(messagesBlockHeight, messagesView, block)
	messages[1] = senderMessageFactory.CreatePrepareMessage(messagesBlockHeight, messagesView, block.BlockHash())
	messages[2] = senderMessageFactory.CreateCommitMessage(messagesBlockHeight, messagesView, block.BlockHash())
	messages[3] = senderMessageFactory.CreateViewChangeMessage(messagesBlockHeight, messagesView, nil)
	messages[4] = senderMessageFactory.CreateNewViewMessage(messagesBlockHeight, messagesView, nil, nil, block)
	allPublicKeys := net.Discovery.AllGossipsPublicKeys()

	return &harness{
		ctx:             ctx,
		net:             net,
		senderNodeIndex: senderNodeIndex,
		senderGossip:    senderGossip,
		block:           block,
		messages:        messages,
		allPublicKeys:   allPublicKeys,
	}
}

func (h *harness) ExpectXMessagesToBeSent(times int) {
	h.senderGossip.When("SendMessage", mock.Any, mock.Any, mock.Any).Times(times)
}

func (h *harness) ExpectEachMessageToBeReceivedXTimes(times int, excludedNodes []int) {

	for index, node := range h.net.Nodes {
		if thisNodeCannotReceiveMessages(index, excludedNodes) {
			continue
		}

		node.Filter.Receiver.(*builders.MockMessageReceiver).When("OnReceivePreprepare", mock.Any, mock.Any).Return().Times(times)
		node.Filter.Receiver.(*builders.MockMessageReceiver).When("OnReceivePrepare", mock.Any, mock.Any).Return().Times(times)
		node.Filter.Receiver.(*builders.MockMessageReceiver).When("OnReceiveCommit", mock.Any, mock.Any).Return().Times(times)
		node.Filter.Receiver.(*builders.MockMessageReceiver).When("OnReceiveViewChange", mock.Any, mock.Any).Return().Times(times)
		node.Filter.Receiver.(*builders.MockMessageReceiver).When("OnReceiveNewView", mock.Any, mock.Any).Return().Times(times)
	}
}

func thisNodeCannotReceiveMessages(index int, excludedNodes []int) bool {

	if len(excludedNodes) == 0 {
		return false
	}

	for _, n := range excludedNodes {
		if index == n {
			return true
		}
	}
	return false
}

func (h *harness) SendAllMessages() error {
	senderGossip := h.net.GetNodeGossip(h.net.Nodes[h.senderNodeIndex].KeyManager.MyPublicKey())
	for _, msg := range h.messages {
		rawMsg := msg.ToConsensusRawMessage()
		if err := senderGossip.SendMessage(h.ctx, h.allPublicKeys, rawMsg); err != nil {
			return err
		}
	}
	return nil
}
func (h *harness) Verify() (bool, error) {
	for _, node := range h.net.Nodes {
		if ok, err := node.Filter.Receiver.(*builders.MockMessageReceiver).Verify(); !ok {
			return ok, err
		}
	}

	return h.senderGossip.Verify()
}

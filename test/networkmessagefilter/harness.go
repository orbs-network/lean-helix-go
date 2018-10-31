package networkmessagefilter

import (
	"context"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
)

const NODE_COUNT = 4

type harness struct {
	ctx          context.Context
	net          *builders.TestNetwork
	receiver     *builders.MockMessageReceiver
	senderNode   *builders.Node
	receiverNode *builders.Node
	messages     []leanhelix.ConsensusMessage
}

func NewHarness(blockHeight primitives.BlockHeight) *harness {

	ctx := context.Background()

	net := builders.ATestNetwork(NODE_COUNT, nil)
	receiverNode := net.Nodes[0]
	senderNode := net.Nodes[1]

	filter := leanhelix.NewNetworkMessageFilter(receiverNode.Gossip, receiverNode.PublicKey)
	mockReceiver := builders.NewMockMessageReceiver()
	filter.SetBlockHeight(ctx, blockHeight, mockReceiver)

	return &harness{
		ctx:          ctx,
		net:          net,
		receiver:     mockReceiver,
		senderNode:   senderNode,
		receiverNode: receiverNode,
	}
}

func (h *harness) GenerateMessages(
	messagesBlockHeight primitives.BlockHeight,
	keyManager leanhelix.KeyManager) {

	block := builders.CreateBlock(builders.GenesisBlock)

	messages := make([]leanhelix.ConsensusMessage, 5)
	messagesView := primitives.View(3)
	messages[0] = builders.APrepreparMessage(keyManager, messagesBlockHeight, messagesView, block)
	messages[1] = builders.APrepareMessage(keyManager, messagesBlockHeight, messagesView, block)
	messages[2] = builders.ACommitMessage(keyManager, messagesBlockHeight, messagesView, block)
	messages[3] = builders.AViewChangeMessage(keyManager, messagesBlockHeight, messagesView, nil)
	messages[4] = builders.ANewViewMessage(keyManager, messagesBlockHeight, messagesView, nil, nil, block)

	h.messages = messages
}

func (h *harness) ExpectEachMessageToBeReceivedXTimes(times int) {
	h.receiver.When("OnReceivePreprepare", mock.Any, mock.Any).Return().Times(times)
	h.receiver.When("OnReceivePrepare", mock.Any, mock.Any).Return().Times(times)
	h.receiver.When("OnReceiveCommit", mock.Any, mock.Any).Return().Times(times)
	h.receiver.When("OnReceiveViewChange", mock.Any, mock.Any).Return().Times(times)
	h.receiver.When("OnReceiveNewView", mock.Any, mock.Any).Return().Times(times)
}

func (h *harness) SendAllMessages() {
	networkCommunication := h.senderNode.Gossip
	allPublicKeys := h.net.Discovery.AllGossipsPublicKeys()
	for _, msg := range h.messages {
		rawMsg := msg.ToConsensusRawMessage()
		networkCommunication.SendMessage(h.ctx, allPublicKeys, rawMsg)
	}
}

func (h *harness) Verify() (bool, error) {
	return h.receiver.Verify()
}

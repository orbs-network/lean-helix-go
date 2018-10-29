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

	net := builders.NewSimpleTestNetwork(NODE_COUNT, blockHeight, nil)
	receiverNode := net.Nodes[0]
	senderNode := net.Nodes[1]

	filter := leanhelix.NewNetworkMessageFilter(receiverNode.Config.NetworkCommunication, receiverNode.Config.KeyManager.MyPublicKey())
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

	senderMessageFactory := leanhelix.NewMessageFactory(keyManager)
	block := builders.CreateBlock(builders.GenesisBlock)

	messages := make([]leanhelix.ConsensusMessage, 5)
	messagesView := primitives.View(3)
	messages[0] = senderMessageFactory.CreatePreprepareMessage(messagesBlockHeight, messagesView, block)
	messages[1] = senderMessageFactory.CreatePrepareMessage(messagesBlockHeight, messagesView, block.BlockHash())
	messages[2] = senderMessageFactory.CreateCommitMessage(messagesBlockHeight, messagesView, block.BlockHash())
	messages[3] = senderMessageFactory.CreateViewChangeMessage(messagesBlockHeight, messagesView, nil)
	messages[4] = senderMessageFactory.CreateNewViewMessage(messagesBlockHeight, messagesView, nil, nil, block)

	h.messages = messages
}

func (h *harness) ExpectEachMessageToBeReceivedXTimes(times int) {
	h.receiver.When("OnReceivePreprepare", mock.Any, mock.Any).Return().Times(times)
	h.receiver.When("OnReceivePrepare", mock.Any, mock.Any).Return().Times(times)
	h.receiver.When("OnReceiveCommit", mock.Any, mock.Any).Return().Times(times)
	h.receiver.When("OnReceiveViewChange", mock.Any, mock.Any).Return().Times(times)
	h.receiver.When("OnReceiveNewView", mock.Any, mock.Any).Return().Times(times)
}

func (h *harness) SendAllMessages() error {
	networkCommunication := h.senderNode.Config.NetworkCommunication
	allPublicKeys := h.net.Discovery.AllGossipsPublicKeys()
	for _, msg := range h.messages {
		rawMsg := msg.ToConsensusRawMessage()
		if err := networkCommunication.SendMessage(h.ctx, allPublicKeys, rawMsg); err != nil {
			return err
		}
	}
	return nil
}

func (h *harness) Verify() (bool, error) {
	return h.receiver.Verify()
}

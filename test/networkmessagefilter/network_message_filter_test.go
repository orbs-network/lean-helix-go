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

	//
	//ctx := context.Background()
	//
	//height := primitives.BlockHeight(3)
	//view := primitives.View(0)
	//net := builders.NewSimpleTestNetwork(NODE_COUNT, nil) // Node 0 is leader
	//
	//senderNode := net.Nodes[0]
	//messageFactory := leanhelix.NewMessageFactory(senderNode.KeyManager)
	//mockReceiver := builders.NewMockMessageReceiver()
	//gossip := net.GetNodeGossip(senderNode.KeyManager.MyPublicKey())
	//
	//mockReceiver.When("OnReceivePreprepare", mock.Any, mock.Any).Times(0)
	//mockReceiver.When("OnReceivePrepare", mock.Any, mock.Any).Times(0)
	//mockReceiver.When("OnReceiveCommit", mock.Any, mock.Any).Times(0)
	//mockReceiver.When("OnReceiveViewChange", mock.Any, mock.Any).Times(0)
	//mockReceiver.When("OnReceiveNewView", mock.Any, mock.Any).Times(0)
	//gossip.When("SendMessage", mock.Any, mock.Any).Times(5)
	//
	//filter := leanhelix.NewNetworkMessageFilter(gossip, height, senderNode.KeyManager.MyPublicKey(), mockReceiver)
	//filter.SetBlockHeight(ctx, 3)
	//
	//block := builders.CreateBlock(builders.GenesisBlock)
	//allNetworkPublicKeys := net.Discovery.AllGossipsPublicKeys()
	//
	//messages := make([]leanhelix.ConsensusMessage, 5)
	//messages[0] = messageFactory.CreatePreprepareMessage(height, view, block)
	//messages[1] = messageFactory.CreatePrepareMessage(height, view, block.BlockHash())
	//messages[2] = messageFactory.CreateCommitMessage(height, view, block.BlockHash())
	//messages[3] = messageFactory.CreateViewChangeMessage(height, view, nil)
	//messages[4] = messageFactory.CreateNewViewMessage(height, view, nil, nil, block)
	//
	//for _, msg := range messages {
	//	rawMsg := msg.ToConsensusRawMessage()
	//	if err := gossip.SendMessage(ctx, allNetworkPublicKeys, rawMsg); err != nil {
	//		t.Error(err)
	//	}
	//}
	//ok, err := mockReceiver.Verify()
	//if !ok {
	//	t.Error(err)
	//}

}

func TestIgnoreMessagesFromNodesNotPartOfTheNetwork(t *testing.T) {

	t.Skip()

	messagesBlockHeight := primitives.BlockHeight(3)
	filterBlockHeight := primitives.BlockHeight(3)
	nonMemberSenderNodeIndex := 2
	nonMemberNodeIndices := []int{nonMemberSenderNodeIndex}

	h := NewHarness(NODE_COUNT, nonMemberSenderNodeIndex, messagesBlockHeight, filterBlockHeight, primitives.View(0), nonMemberNodeIndices)
	h.ExpectEachMessageToBeReceivedXTimes(0, []int{nonMemberSenderNodeIndex})
	h.ExpectXMessagesToBeSent(5)

}

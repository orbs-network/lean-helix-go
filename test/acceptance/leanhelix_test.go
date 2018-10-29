package acceptance

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/gossip"
	"github.com/stretchr/testify/require"
	"testing"
)

// Adapted from PBFT.spec.ts

// Some terminology:
// These tests are as close to e2e tests as a library can get.
// - e2e tests exercise the whole system but there is no system here (again this being a library)
// - Acceptance tests are similar to e2e tests except that they don't set up any
// time-consuming components (so no network, no disk I/O)
// - Component tests exercise a single component and mock everything around it.
// It you call this library a component then it is a component test.

// TODO - decide if it's a single component or multiple components (in which case this file contains acceptance tests)
// TODO: Use TestSyncCompletePetitionerSyncFlow for inspiration (When(), expect(), EventuallyVerify(), VerifyMocks() etc)

const NODE_COUNT = 4

func TestSendPreprepareOnlyIfLeader(t *testing.T) {

	// TODO This is a big test, it is not ready yet - it is pending some LeanHelixTerm tests to pass

	t.Skip()

	net := builders.NewSimpleTestNetwork(NODE_COUNT, 0, nil) // Node 0 is leader

	predicateMessageTypeIsPreprepare := func(msg interface{}) bool {
		message := msg.(lh.ConsensusMessage)
		return message.MessageType() == lh.LEAN_HELIX_PREPREPARE
	}

	gossips := make([]*gossip.Gossip, 0, len(net.Nodes))
	for i := range net.Nodes {
		gossip := net.GetNodeGossip(net.Nodes[i].Config.KeyManager.MyPublicKey())
		if gossip == nil {
			t.Errorf("Cannot find Gossip for node #%v: %v", i, net.Nodes[i].Config.KeyManager.MyPublicKey())
		}
		gossips = append(gossips, gossip)
	}

	defer net.Stop()
	err := net.StartConsensusOnAllNodes()
	if err != nil {
		t.Error(err)
	}

	net.BlockUtils.ProvideNextBlock()
	net.BlockUtils.ResolveAllValidations(true)

	require.Equal(t, 1, gossips[0].StatsNumSentMessages(predicateMessageTypeIsPreprepare), "node 0 did not send 1 preprepare")
	require.Equal(t, 0, gossips[1].StatsNumSentMessages(predicateMessageTypeIsPreprepare), "node 1 did not send 1 preprepare")
	require.Equal(t, 0, gossips[2].StatsNumSentMessages(predicateMessageTypeIsPreprepare), "node 2 did not send 1 preprepare")
	require.Equal(t, 0, gossips[3].StatsNumSentMessages(predicateMessageTypeIsPreprepare), "node 3 did not send 1 preprepare")

	// 17-Sep-2018 plan for this test
	// Code the harness to generate the test network ("Network")

	// Create instance of the Harness
	// Each node has: real LeanHelix, mock BlockUtils, mock KeyManager, mock NetworkCommunication, Config?
	// StartConsensusOnAllNodes the Network - each node starts the infinite loop and listens on messages
	// Spy (hook) on all messages that are send by the nodes
	// Provide the next block from outside the Network
	// Verify that only the leader sends out a Preprepare message
	// Graceful shutdown

	// 20-SEP-2018 plan:
	// Write the tests as-is, refactor to harness later.

	//h := NewHarness(NODE_COUNT)
	//h.expectLeaderToSendPreprepareMessageOnce()
	//h.expectNonLeaderToNotSendPreprepareMessage()
	//
	//h.Verify()

	//
	//h.service.network.When("sendToMembers", &services.CommitBlockInput{expectedBlockPair}).Return(nil, nil).Times(1)
	//h.service.network.When("SendBenchmarkConsensusCommitted", mock.AnyIf(fmt.Sprintf("LastCommittedBlockHeight equals %d, recipient equals %s and sender equals %s", expectedLastCommitted, expectedRecipient, expectedSender), lastCommittedReplyMatcher)).Times(1)
	//h.service.StartConsensusOnAllNodes()
	//defer h.service.Stop()

}

/*
   const { testNetwork, blockUtils } = aSimpleTestNetwork();
   const node0 = testNetwork.nodes[0];
   const node1 = testNetwork.nodes[1];
   const node2 = testNetwork.nodes[2];
   const node3 = testNetwork.nodes[3];
   const gossip0 = testNetwork.getNodeGossip(node0.pk);
   const gossip1 = testNetwork.getNodeGossip(node1.pk);
   const gossip2 = testNetwork.getNodeGossip(node2.pk);
   const gossip3 = testNetwork.getNodeGossip(node3.pk);
   const spy0 = sinon.spy(gossip0, "multicast");
   const spy1 = sinon.spy(gossip1, "multicast");
   const spy2 = sinon.spy(gossip2, "multicast");
   const spy3 = sinon.spy(gossip3, "multicast");

   testNetwork.startConsensusOnAllNodes();
   await nextTick();
   await blockUtils.provideNextBlock();
   await blockUtils.resolveAllValidations(true);
   await nextTick(); // await for notifyCommitted
   const preprepareCounter = (spy: sinon.SinonSpy) => {
       return spy.getCalls().filter(c => c.args[1].content.messageType === MessageType.PREPREPARE).length;
   };

   expect(preprepareCounter(spy0)).to.equal(1);
   expect(preprepareCounter(spy1)).to.equal(0);
   expect(preprepareCounter(spy2)).to.equal(0);
   expect(preprepareCounter(spy3)).to.equal(0);

   testNetwork.shutDown();

*/

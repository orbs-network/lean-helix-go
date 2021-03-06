package acceptance

import (
	"context"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/leaderelection"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"sync"
	"testing"
	"time"
)

const TIMEOUT = 1 * time.Second

type SimpleMockBlockUtils struct {
	mock.Mock
}

func (b *SimpleMockBlockUtils) RequestNewBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, _ primitives.MemberId, _ interfaces.Block) (interfaces.Block, primitives.BlockHash) {
	res := b.Called(ctx, blockHeight) // go-mock chokes on nil values, we don't need it, so don't pass it
	return res.Get(0).(interfaces.Block), res.Get(1).(primitives.BlockHash)
}

func (b SimpleMockBlockUtils) ValidateBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, _ primitives.MemberId, block interfaces.Block, blockHash primitives.BlockHash, prevBlock interfaces.Block) error {
	return b.Called(ctx, blockHeight, block, blockHash, prevBlock).Error(0)
}

func (b SimpleMockBlockUtils) ValidateBlockCommitment(blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash) bool {
	return b.Called(blockHeight, block, blockHash).Bool(0)
}

func TestNodeSyncIsStillHandledDespiteBlockedOnRequestNewBlockProposal(t *testing.T) {

	// Set to pause on RequestNewBlockProposal of H=1 and verify it has paused
	// Call sync with a valid block with H=2
	// Verify node0 reached H=3
	// Set to pause on RequestNewBlockProposal again and verify it has paused

	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		block3 := mocks.ABlock(block2)

		net := network.ATestNetworkBuilder(4, block1, block2, block3).
			//LogToConsole(t).
			Build(ctx)
		node0 := net.Nodes[0]

		net.SetNodesToPauseOnRequestNewBlock()
		net.StartConsensus(ctx)
		net.ReturnWhenNodeIsPausedOnRequestNewBlock(ctx, node0)
		bc, err := leaderelection.GenerateBlocksWithProofsForTest([]interfaces.Block{block1, block2, block3}, net.Nodes)
		if err != nil {
			t.Fatalf("Error creating mock blockchain for tests - %s", err)
			return
		}
		blockToSync, blockProofToSync := bc.BlockAndProofAt(2)
		prevBlockToSync, prevBlockProofToSync := bc.BlockAndProofAt(1)
		if err := node0.Sync(ctx, blockToSync, blockProofToSync, prevBlockToSync, prevBlockProofToSync); err != nil {
			t.Fatalf("Sync failed for node %s - %s", node0.MemberId, err)
		}
		net.WaitUntilSubsetOfNodesEventuallyReachASpecificHeight(ctx, 3, 1)
	})
}

func TestRequestNewBlockDoesNotHangElectionsTrigger(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		withConsensusRound(ctx, t, func(net *network.TestNetwork, blockUtilsMocks []*SimpleMockBlockUtils, blockToPropose interfaces.Block) {
			node0 := net.Nodes[0]
			blockUtilsMock := blockUtilsMocks[0]

			createNewBlockProposalEntered := newWaitingGroupWithDelta(1)
			createNewBlockProposalCanceled := newWaitingGroupWithDelta(1)
			blockUtilsMock.
				When("RequestNewBlockProposal", mock.Any, mock.Any).
				Call(func(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock interfaces.Block) (interfaces.Block, primitives.BlockHash) {
					createNewBlockProposalEntered.Done()
					<-ctx.Done() // block until context cancellation
					createNewBlockProposalCanceled.Done()
					return blockToPropose, nil
				})
			for _, b := range blockUtilsMocks {
				b.When("ValidateBlockCommitment", mock.Any, mock.Any, mock.Any).Return(true)
			}
			node0.StartConsensus(ctx)

			createNewBlockProposalEntered.Wait()

			electionsTriggerProcessed := newWaitingGroupWithDelta(1)
			go func() {
				<-node0.TriggerElectionOnNode(ctx)
				electionsTriggerProcessed.Done()
			}()

			test.FailIfNotDoneByTimeout(t, electionsTriggerProcessed, TIMEOUT, "Election trigger is blocked by RequestNewBlockProposal")
			test.FailIfNotDoneByTimeout(t, createNewBlockProposalCanceled, TIMEOUT, "RequestNewBlockProposal's ctx was not canceled immediately after election trigger")
		})
	})
}

func newWaitingGroupWithDelta(delta int) *sync.WaitGroup {
	createNewBlockProposalEntered := sync.WaitGroup{}
	createNewBlockProposalEntered.Add(delta)
	return &createNewBlockProposalEntered
}

func withConsensusRound(ctx context.Context, tb testing.TB, test func(net *network.TestNetwork, blockUtilsMock []*SimpleMockBlockUtils, blockToPropose interfaces.Block)) {
	nodeCount := 4

	block1 := mocks.ABlock(interfaces.GenesisBlock)
	//instanceId := primitives.InstanceId(rand.Uint64())

	var simpleMockBlockUtils []*SimpleMockBlockUtils
	var blockUtils []interfaces.BlockUtils
	for i := 0; i < nodeCount; i++ {
		aBlockUtils := &SimpleMockBlockUtils{}
		simpleMockBlockUtils = append(simpleMockBlockUtils, aBlockUtils)
		blockUtils = append(blockUtils, aBlockUtils)
	}
	net := network.NewTestNetworkBuilder().
		WithNodeCount(nodeCount).
		WithBlockUtils(blockUtils).
		//InNetwork(instanceId).
		//LogToConsole(tb).
		Build(ctx)

	test(net, simpleMockBlockUtils, block1)
}

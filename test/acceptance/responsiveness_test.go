package acceptance

import (
	"context"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/orbs-network/lean-helix-go/test/network"
	"math/rand"
	"sync"
	"testing"
	"time"
)

const TIMEOUT = 1 * time.Second

// TODO -  a workaround for a bug in go-mock. when passing nil interface type to Call() implementation - Mock.Called() fails to invoke the Call function.
type nilBlock struct{}

func (nb *nilBlock) Height() primitives.BlockHeight {
	panic("I'm a mock object for a nil value and this would throw nil pointer exception")
}

type SimpleMockBlockUtils struct {
	mock.Mock
}

func (b *SimpleMockBlockUtils) RequestNewBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock interfaces.Block) (interfaces.Block, primitives.BlockHash) {
	if prevBlock == nil {
		prevBlock = &nilBlock{} // mock object cannot handle nil interfaces
	}
	res := b.Called(ctx, blockHeight, prevBlock)
	return res.Get(0).(interfaces.Block), res.Get(1).(primitives.BlockHash)
}

func (b SimpleMockBlockUtils) ValidateBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash, prevBlock interfaces.Block) error {
	return b.Called(ctx, blockHeight, block, blockHash, prevBlock).Error(0)
}

func (b SimpleMockBlockUtils) ValidateBlockCommitment(blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash) bool {
	return b.Called(blockHeight, block, blockHash).Bool(0)
}

func TestRequestNewBlockDoesNotHangNodeSync(t *testing.T) {
	t.Skip() // TODO - remove skip when worker-go-routine is implemented!!
	test.WithContext(func(ctx context.Context) {
		withConsensusRound(func(net *network.TestNetwork, blockUtilsMock *SimpleMockBlockUtils, blockToPropose interfaces.Block) {
			node0 := net.Nodes[0]

			createNewBlockProposalEntered := newWaitingGroupWithDelta(1)
			createNewBlockProposalCancelled := newWaitingGroupWithDelta(1)
			blockUtilsMock.
				When("RequestNewBlockProposal", mock.Any, mock.Any, mock.Any).
				Call(func(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock interfaces.Block) (interfaces.Block, primitives.BlockHash) {
					createNewBlockProposalEntered.Done()
					<-ctx.Done() // block until context cancellation
					createNewBlockProposalCancelled.Done()
					return blockToPropose, nil
				})

			node0.StartConsensus(ctx)

			createNewBlockProposalEntered.Wait()

			updateStateCompleted := newWaitingGroupWithDelta(1)
			go func() {
				node0.SyncWithoutProof(ctx, nil, nil)
				updateStateCompleted.Done()
			}()

			requireDone(t, updateStateCompleted, 1*time.Second, "NodeSync is blocked by RequestNewBlockProposal")
			requireDone(t, createNewBlockProposalCancelled, 1*time.Second, "RequestNewBlockProposal's ctx was not cancelled immediately after NodeSync")
		})

	})
}

func TestRequestNewBlockDoesNotHangElectionsTrigger(t *testing.T) {
	t.Skip() // TODO - remove skip when worker-go-routine is implemented!!
	test.WithContext(func(ctx context.Context) {
		withConsensusRound(func(net *network.TestNetwork, blockUtilsMock *SimpleMockBlockUtils, blockToPropose interfaces.Block) {
			node0 := net.Nodes[0]

			createNewBlockProposalEntered := newWaitingGroupWithDelta(1)
			createNewBlockProposalCancelled := newWaitingGroupWithDelta(1)
			blockUtilsMock.
				When("RequestNewBlockProposal", mock.Any, mock.Any, mock.Any).
				Call(func(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock interfaces.Block) (interfaces.Block, primitives.BlockHash) {
					createNewBlockProposalEntered.Done()
					<-ctx.Done() // block until context cancellation
					createNewBlockProposalCancelled.Done()
					return blockToPropose, nil
				})

			node0.StartConsensus(ctx)

			createNewBlockProposalEntered.Wait()

			electionsTriggerProcessed := newWaitingGroupWithDelta(1)
			go func() {
				<-node0.TriggerElection(ctx)
				electionsTriggerProcessed.Done()
			}()

			requireDone(t, electionsTriggerProcessed, TIMEOUT, "Election trigger is blocked by RequestNewBlockProposal")
			requireDone(t, createNewBlockProposalCancelled, TIMEOUT, "RequestNewBlockProposal's ctx was not cancelled immediately after election trigger")
		})
	})
}

func newWaitingGroupWithDelta(delta int) *sync.WaitGroup {
	createNewBlockProposalEntered := sync.WaitGroup{}
	createNewBlockProposalEntered.Add(delta)
	return &createNewBlockProposalEntered
}

func withConsensusRound(test func (net *network.TestNetwork, blockUtilsMock *SimpleMockBlockUtils, blockToPropose interfaces.Block)) {
	block1 := mocks.ABlock(interfaces.GenesisBlock)
	block2 := mocks.ABlock(block1)
	instanceId := primitives.InstanceId(rand.Uint64())
	mockBlockUtils := &SimpleMockBlockUtils{}
	net := network.NewTestNetworkBuilder().
		WithNodeCount(4).
		WithBlocks([]interfaces.Block{block1, block2}).
		WithBlockUtils(mockBlockUtils).
		InNetwork(instanceId).
		LogToConsole().
		Build()

	test(net, mockBlockUtils, block1)
}

func requireDone(t *testing.T, waitGroup *sync.WaitGroup, timeout time.Duration, format string, args ...interface{}) {
	timeoutCtx, _ := context.WithTimeout(context.Background(), timeout)

	condDone := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(condDone)
	}()

	select {
	case <-condDone: // wait group finished waiting
	case <-timeoutCtx.Done():
		t.Fatalf(format, args...)
	}
}
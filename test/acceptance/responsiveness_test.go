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
	"testing"
	"time"
)

type SimpleMockBlockUtils struct {
	mock.Mock
}

func (b *SimpleMockBlockUtils) GetValidationResult() bool {
	panic("implement me")
}

func (b *SimpleMockBlockUtils) SetValidationResult(bool) {
	panic("implement me")
}

type nilBlock struct{}

func (nb *nilBlock) Height() primitives.BlockHeight {
	panic("I'm a mock object for a nil value and this would through nil pointer exception")
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
	t.Skip() // FIXME: Fails
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		//net := network.ATestNetwork(4, block1, block2)

		instanceId := primitives.InstanceId(rand.Uint64())
		mockBlockUtils := &SimpleMockBlockUtils{}

		net := network.NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]interfaces.Block{block1, block2}).
			WithBlockUtils(mockBlockUtils).
			InNetwork(instanceId).
			LogToConsole().
			Build()

		node0 := net.Nodes[0]

		// from harness - get mock for BlockUtils.CreateNewBlockProposal - like so:
		createNewBlockProposalEntered := make(chan struct{})
		createNewBlockProposalCompleted := make(chan struct{})
		mockBlockUtils.
			When("RequestNewBlockProposal", mock.Any, mock.Any, mock.Any).
			Call(func(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock interfaces.Block) (interfaces.Block, primitives.BlockHash) {
				close(createNewBlockProposalEntered)
				<-ctx.Done()
				close(createNewBlockProposalCompleted)
				return block1, nil
			})

		// wait for its childCtx which will be cancelled when NodeSync is called

		//net.SetNodesToPauseOnRequestNewBlock()

		node0.StartConsensus(ctx)
		<-createNewBlockProposalEntered // this assures CreateNewBlockProposal is underway

		//net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node0)

		//net.SetNodesToPauseOnHandleUpdateState()

		updateStateCompleted := make(chan struct{})
		go func() {
			node0.SyncWithoutProof(ctx, nil, nil)
			updateStateCompleted <- struct{}{}
		}()

		requireChanWriteWithinTimeout(t, updateStateCompleted, 1*time.Second, "NodeSync is blocked by RequestNewBlockProposal")
		requireChanWriteWithinTimeout(t, createNewBlockProposalCompleted, 1*time.Second, "createNewBlockProposal did not return immediately after NodeSync")

		// net.ResumeRequestNewBlockOnNodes(ctx, node0)

		//net.ReturnWhenNodesPauseOnUpdateState(ctx, node0)
	})
}

func requireChanWriteWithinTimeout(t *testing.T, listenChan <-chan struct{}, timeout time.Duration, format string, args ...interface{}) {
	timeoutCtx, _ := context.WithTimeout(context.Background(), timeout)
	select {
	case <-listenChan: // the event we are anticipating
	case <-timeoutCtx.Done():
		t.Fatalf(format, args...)
	}
}

package acceptance

import (
	"context"
	"fmt"
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

type MockBlockUtils struct {
	mock.Mock
}

func (b *MockBlockUtils) RequestNewBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock interfaces.Block) (interfaces.Block, primitives.BlockHash) {
	res := b.Called(ctx, blockHeight, prevBlock)
	return res.Get(0).(interfaces.Block), res.Get(1).(primitives.BlockHash)
}

func (b MockBlockUtils) ValidateBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash, prevBlock interfaces.Block) error {
	return b.Called(ctx, blockHeight, block, blockHash, prevBlock).Error(0)
}

func (b MockBlockUtils) ValidateBlockCommitment(blockHeight primitives.BlockHeight, block interfaces.Block, blockHash primitives.BlockHash) bool {
	return b.Called(blockHeight, block, blockHash).Bool(0)
}

func TestRequestNewBlockDoesNotHangNodeSync(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		block1 := mocks.ABlock(interfaces.GenesisBlock)
		block2 := mocks.ABlock(block1)
		//net := network.ATestNetwork(4, block1, block2)

		instanceId := primitives.InstanceId(rand.Uint64())
		mockBlockUtils := &MockBlockUtils{}
		fmt.Println("Bla 1")

		net := network.NewTestNetworkBuilder().
			WithNodeCount(4).
			WithBlocks([]interfaces.Block{block1, block2}).
			WithBlockUtils(mockBlockUtils).
			InNetwork(instanceId).
			LogToConsole().
			Build()

		node0 := net.Nodes[0]
		fmt.Println("Bla 2")
		// from harness - get mock for BlockUtils.CreateNewBlockProposal - like so:
		createNewBlockProposalEntered := make(chan struct{})
		createNewBlockProposalExited := make(chan struct{})
		mockBlockUtils.
			When("RequestNewBlockProposal", mock.Any, mock.Any, mock.Any).
			Call(func(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock interfaces.Block) (interfaces.Block, primitives.BlockHash) {
				fmt.Println("Bla inside ")

				close(createNewBlockProposalEntered)
				<-ctx.Done()
				close(createNewBlockProposalExited)
				return block1, nil
			})
		fmt.Println("Bla 3")

		// wait for its childCtx which will be cancelled when NodeSync is called

		//net.SetNodesToPauseOnRequestNewBlock()

		node0.StartConsensus(ctx)
		fmt.Println("Bla 4")

		<-createNewBlockProposalEntered // this assures CreateNewBlockProposal is underway

		//net.ReturnWhenNodesPauseOnRequestNewBlock(ctx, node0)

		fmt.Println("Bla 5")

		//net.SetNodesToPauseOnHandleUpdateState()

		doneNodeSync := make(chan struct{})
		timeoutCtx, _ := context.WithTimeout(ctx, 1*time.Second)
		go func() {
			fmt.Println("Bla goooooo")
			node0.SyncWithoutProof(ctx, nil, nil)
			doneNodeSync <- struct{}{}
		}()

		select {
		case <-doneNodeSync:
			t.Log("NodeSync finished successfully")

		case <-timeoutCtx.Done():
			t.Errorf("Timed out waiting for NodeSync")
		}

		select {
		case <-createNewBlockProposalExited:
			t.Log("createNewBlockProposal terminated as expected")

		case <-timeoutCtx.Done():
			t.Errorf("Timed out waiting for createNewBlockProposal context to be cancelled")
		}
		// net.ResumeRequestNewBlockOnNodes(ctx, node0)

		//net.ReturnWhenNodesPauseOnUpdateState(ctx, node0)
	})
}

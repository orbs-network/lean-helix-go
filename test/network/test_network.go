package network

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/matchers"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

type TestNetwork struct {
	NetworkId primitives.NetworkId
	Nodes     []*Node
	Discovery *mocks.Discovery
}

func (net *TestNetwork) GetNodeCommunication(memberId primitives.MemberId) *mocks.CommunicationMock {
	return net.Discovery.GetCommunicationById(memberId)
}

func (net *TestNetwork) TriggerElection(ctx context.Context) {
	for _, node := range net.Nodes {
		node.TriggerElection(ctx)
	}
}

func (net *TestNetwork) StartConsensus(ctx context.Context) *TestNetwork {
	for _, node := range net.Nodes {
		node.StartConsensus(ctx)
	}

	return net
}

func (net *TestNetwork) StartConsensusSync(ctx context.Context) *TestNetwork {
	for _, node := range net.Nodes {
		node.StartConsensusSync(ctx)
	}

	return net
}

func (net *TestNetwork) RegisterNode(node *Node) {
	net.Nodes = append(net.Nodes, node)
}

func (net *TestNetwork) RegisterNodes(nodes []*Node) {
	for _, node := range nodes {
		net.RegisterNode(node)
	}
}

func (net *TestNetwork) AllNodesChainEndsWithABlock(block interfaces.Block) bool {
	for _, node := range net.Nodes {
		if matchers.BlocksAreEqual(block, node.GetLatestBlock()) == false {
			return false
		}
	}
	return true
}

func (net *TestNetwork) WaitForAllNodesToCommitBlock(ctx context.Context, block interfaces.Block) bool {
	for _, node := range net.Nodes {
		select {
		case <-ctx.Done():
			return false
		case nodeState := <-node.NodeStateChannel:
			if matchers.BlocksAreEqual(block, nodeState.block) == false {
				return false
			}
		}
	}
	return true
}

const MINIMUM_NUMBER_OF_NODES_FOR_CONSENSUS = 4

func (net *TestNetwork) WaitForAllNodesToCommitTheSameBlock(ctx context.Context) bool {
	if len(net.Nodes) < MINIMUM_NUMBER_OF_NODES_FOR_CONSENSUS {
		panic("Not enough nodes for consensus")
	}

	select {
	case <-ctx.Done():
		return false
	case firstNodeStateChannel := <-net.Nodes[0].NodeStateChannel:
		firstNodeBlock := firstNodeStateChannel.block
		for i := 1; i < len(net.Nodes); i++ {
			node := net.Nodes[i]

			select {
			case <-ctx.Done():
				return false

			case nodeState := <-node.NodeStateChannel:
				if matchers.BlocksAreEqual(firstNodeBlock, nodeState.block) == false {
					return false
				}
			}

		}
		return true
	}
}

func (net *TestNetwork) NodesPauseOnValidate(nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}

	for _, node := range nodes {
		node.BlockUtils.PauseOnValidation = true
	}
}

func (net *TestNetwork) WaitForNodesToValidate(ctx context.Context, nodes ...*Node) {
	for _, node := range nodes {
		node.BlockUtils.ValidationSns.WaitForSignal(ctx)
	}
}

func (net *TestNetwork) ResumeNodesValidation(ctx context.Context, nodes ...*Node) {
	for _, node := range nodes {
		node.BlockUtils.ValidationSns.Resume(ctx)
	}
}

func (net *TestNetwork) NodesPauseOnRequestNewBlock(nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}

	for _, node := range nodes {
		node.BlockUtils.PauseOnRequestNewBlock = true
	}
}

func (net *TestNetwork) WaitForNodeToRequestNewBlock(ctx context.Context, node *Node) {
	node.BlockUtils.RequestNewBlockSns.WaitForSignal(ctx)
}

func (net *TestNetwork) ResumeNodeRequestNewBlock(ctx context.Context, node *Node) {
	node.BlockUtils.RequestNewBlockSns.Resume(ctx)
}

func (net *TestNetwork) WaitForConsensus(ctx context.Context) {
	for _, node := range net.Nodes {
		select {
		case <-ctx.Done():
			return
		case <-node.NodeStateChannel:
		}
	}
}

func (net *TestNetwork) WaitForNodesToCommitABlock(ctx context.Context, nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}

	for _, node := range nodes {
		select {
		case <-ctx.Done():
			return
		case <-node.NodeStateChannel:
		}
	}
}

func (net *TestNetwork) WaitForNodesToCommitASpecificBlock(ctx context.Context, block interfaces.Block, nodes ...*Node) bool {
	if nodes == nil {
		nodes = net.Nodes
	}

	for _, node := range nodes {

		select {
		case <-ctx.Done():
			return false
		case nodeState := <-node.NodeStateChannel:
			if matchers.BlocksAreEqual(block, nodeState.block) == false {
				return false
			}
		}

	}
	return true
}

func (net *TestNetwork) AllNodesValidatedNoMoreThanOnceBeforeCommit(ctx context.Context) bool {
	for _, node := range net.Nodes {

		select {
		case <-ctx.Done():
			return false
		case nodeState := <-node.NodeStateChannel:
			if nodeState.validationCount > 1 {
				return false
			}
		}
	}
	return true
}

func NewTestNetwork(networkId primitives.NetworkId, discovery *mocks.Discovery) *TestNetwork {
	return &TestNetwork{
		NetworkId: networkId,
		Nodes:     []*Node{},
		Discovery: discovery,
	}
}

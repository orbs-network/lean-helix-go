package builders

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type TestNetwork struct {
	Nodes     []*Node
	Discovery *gossip.Discovery
}

func (net *TestNetwork) GetNodeGossip(memberId primitives.MemberId) *gossip.Gossip {
	return net.Discovery.GetGossipById(memberId)
}

func (net *TestNetwork) TriggerElection() {
	for _, node := range net.Nodes {
		node.TriggerElection()
	}
}

func (net *TestNetwork) StartConsensus(ctx context.Context) *TestNetwork {
	for _, node := range net.Nodes {
		node.StartConsensus(ctx)
	}

	return net
}

func (net *TestNetwork) StartConsensusSync() *TestNetwork {
	for _, node := range net.Nodes {
		node.StartConsensusSync()
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

func (net *TestNetwork) ShutDown() {
	// TODO: Is this needed?
	//for _, node := range net.Nodes {
	//	node.Dispose()
	//}
}

func (net *TestNetwork) AllNodesChainEndsWithABlock(block leanhelix.Block) bool {
	for _, node := range net.Nodes {
		if BlocksAreEqual(block, node.GetLatestBlock()) == false {
			return false
		}
	}
	return true
}

func (net *TestNetwork) WaitForAllNodesToCommitBlock(block leanhelix.Block) bool {
	for _, node := range net.Nodes {
		nodeState := <-node.NodeStateChannel
		if BlocksAreEqual(block, nodeState.block) == false {
			return false
		}
	}
	return true
}

const MINIMUM_NUMBER_OF_NODES_FOR_CONSENSUS = 4

func (net *TestNetwork) WaitForAllNodesToCommitTheSameBlock() bool {
	if len(net.Nodes) < MINIMUM_NUMBER_OF_NODES_FOR_CONSENSUS {
		panic("Not enough nodes for consensus")
	}

	firstNodeStateChannel := <-net.Nodes[0].NodeStateChannel
	firstNodeBlock := firstNodeStateChannel.block
	for i := 1; i < len(net.Nodes); i++ {
		node := net.Nodes[i]
		nodeState := <-node.NodeStateChannel
		if BlocksAreEqual(firstNodeBlock, nodeState.block) == false {
			return false
		}
	}
	return true
}

func (net *TestNetwork) NodesPauseOnValidate(nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}

	for _, node := range nodes {
		node.BlockUtils.PauseOnValidation = true
	}
}

func (net *TestNetwork) WaitForNodesToValidate(nodes ...*Node) {
	for _, node := range nodes {
		node.BlockUtils.ValidationSns.WaitForSignal()
	}
}

func (net *TestNetwork) ResumeNodesValidation(nodes ...*Node) {
	for _, node := range nodes {
		node.BlockUtils.ValidationSns.Resume()
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

func (net *TestNetwork) WaitForNodeToRequestNewBlock(node *Node) {
	node.BlockUtils.RequestNewBlockSns.WaitForSignal()
}

func (net *TestNetwork) ResumeNodeRequestNewBlock(node *Node) {
	node.BlockUtils.RequestNewBlockSns.Resume()
}

func (net *TestNetwork) WaitForConsensus() {
	for _, node := range net.Nodes {
		<-node.NodeStateChannel
	}
}

func (net *TestNetwork) WaitForNodesToCommitABlock(nodes ...*Node) {
	if nodes == nil {
		nodes = net.Nodes
	}

	for _, node := range nodes {
		<-node.NodeStateChannel
	}
}

func (net *TestNetwork) WaitForNodesToCommitASpecificBlock(block leanhelix.Block, nodes ...*Node) bool {
	if nodes == nil {
		nodes = net.Nodes
	}

	for _, node := range nodes {
		nodeState := <-node.NodeStateChannel
		if BlocksAreEqual(block, nodeState.block) == false {
			return false
		}
	}
	return true
}

func (net *TestNetwork) AllNodesValidatedNoMoreThanOnceBeforeCommit() bool {
	for _, node := range net.Nodes {
		nodeState := <-node.NodeStateChannel
		if nodeState.validationCount > 1 {
			return false
		}
	}
	return true
}

func NewTestNetwork(discovery *gossip.Discovery) *TestNetwork {
	return &TestNetwork{
		Nodes:     []*Node{},
		Discovery: discovery,
	}
}

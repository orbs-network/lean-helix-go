package builders

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type TestNetwork struct {
	Nodes      []*Node
	BlocksPool []leanhelix.Block
	Discovery  *gossip.Discovery
}

func (net *TestNetwork) GetNodeGossip(publicKey Ed25519PublicKey) *gossip.Gossip {
	return net.Discovery.GetGossipByPK(publicKey)
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

func (net *TestNetwork) AllNodesAgreeOnBlock(block leanhelix.Block) bool {
	for _, node := range net.Nodes {
		nodeState := <-node.NodeStateChannel
		if CalculateBlockHash(block).Equal(CalculateBlockHash(nodeState.block)) == false {
			return false
		}
	}
	return true
}

const MINIMUM_NUMBER_OF_NODES_FOR_CONSENSUS = 4

func (net *TestNetwork) InConsensus() bool {
	if len(net.Nodes) < MINIMUM_NUMBER_OF_NODES_FOR_CONSENSUS {
		panic("Not enough nodes for consensus")
	}

	firstNodeStateChannel := <-net.Nodes[0].NodeStateChannel
	firstNodeBlock := firstNodeStateChannel.block
	for i := 1; i < len(net.Nodes); i++ {
		node := net.Nodes[i]
		nodeState := <-node.NodeStateChannel
		if CalculateBlockHash(firstNodeBlock).Equal(CalculateBlockHash(nodeState.block)) == false {
			return false
		}
	}
	return true
}

func (net *TestNetwork) WaitForConsensus() {
	for _, node := range net.Nodes {
		<-node.NodeStateChannel
	}
}

func (net *TestNetwork) PauseNodesExecutionOnValidation(nodes ...*Node) func() func(isValid bool) {
	for _, node := range nodes {
		node.BlockUtils.PauseOnValidations = true
	}

	return func() func(isValid bool) {
		releasingChannels := make([]chan bool, len(nodes))
		for _, node := range nodes {
			releasingChannel := <-node.BlockUtils.PausingChannel
			releasingChannels = append(releasingChannels, releasingChannel)
		}

		return func(isValid bool) {
			for _, releasingChannel := range releasingChannels {
				releasingChannel <- isValid
			}
		}
	}
}

func (net *TestNetwork) ResolveAllValidations() {
	for _, node := range net.Nodes {
		releasingChannel := make(chan bool)
		node.BlockUtils.PausingChannel <- releasingChannel
	}
}

func (net *TestNetwork) AllNodesValidatedOnceBeforeCommit() bool {
	for _, node := range net.Nodes {
		nodeState := <-node.NodeStateChannel
		if nodeState.validationCount > 1 {
			return false
		}
	}
	return true
}

func NewTestNetwork(discovery *gossip.Discovery, blocksPool []leanhelix.Block) *TestNetwork {
	return &TestNetwork{
		Nodes:      []*Node{},
		BlocksPool: blocksPool,
		Discovery:  discovery,
	}
}

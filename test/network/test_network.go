package network

import "github.com/orbs-network/lean-helix-go/test/builders"

type TestNetwork struct {
	nodes []*builders.Node
}

func NewTestNetwork() *TestNetwork {
	return &TestNetwork{
		nodes: []*builders.Node{},
	}
}

func (testNetwork *TestNetwork) GetNodeGossip() {
	// TODO: Implement
}

func (testNetwork *TestNetwork) RegisterNode(node *builders.Node) {
	testNetwork.nodes = append(testNetwork.nodes, node)
}

func (testNetwork *TestNetwork) RegisterNodes(nodes []*builders.Node) {
	for _, node := range nodes {
		testNetwork.RegisterNode(node)
	}
}

func (testNetwork *TestNetwork) StartConsensusOnAllNodes() {
	for _, node := range testNetwork.nodes {
		node.StartConsensus()
	}
}

func (testNetwork *TestNetwork) Shutdown() {
	for _, node := range testNetwork.nodes {
		node.Dispose()
	}
}

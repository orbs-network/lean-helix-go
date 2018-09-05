package network

type TestNetwork struct {
	nodes []*Node
}

func NewTestNetwork() *TestNetwork {
	return &TestNetwork{
		nodes: []*Node{},
	}
}

func (testNetwork *TestNetwork) GetNodeGossip() {
	// TODO: Implement
}

func (testNetwork *TestNetwork) RegisterNode(node *Node) {
	testNetwork.nodes = append(testNetwork.nodes, node)
}

func (testNetwork *TestNetwork) RegisterNodes(nodes []*Node) {
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

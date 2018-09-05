package builders

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/test/network"
)

type NodeBuilder struct {
	publicKey       lh.PublicKey
	electionTrigger lh.ElectionTrigger
}

func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{}
}

func (nb *NodeBuilder) ElectingLeaderUsing(electionTrigger lh.ElectionTrigger) *NodeBuilder {
	if nb.electionTrigger == nil {
		nb.electionTrigger = electionTrigger
	}
	return nb
}

func (nb *NodeBuilder) WithPK(publicKey lh.PublicKey) *NodeBuilder {
	if nb.publicKey == "" {
		nb.publicKey = publicKey
	}
	return nb
}

func (nb *NodeBuilder) buildConfig() *lh.Config {
	return &lh.Config{
		ElectionTrigger: nb.electionTrigger,
	}
}

func (nb *NodeBuilder) Build() *network.Node {
	return network.NewNode(nb.publicKey, nb.buildConfig())
}

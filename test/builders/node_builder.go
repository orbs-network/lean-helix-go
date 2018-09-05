package builders

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/network"
	"github.com/orbs-network/lean-helix-go/types"
)

type NodeBuilder struct {
	publicKey       types.PublicKey
	electionTrigger leanhelix.ElectionTrigger
}

func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{}
}

func (nb *NodeBuilder) ElectingLeaderUsing(electionTrigger leanhelix.ElectionTrigger) *NodeBuilder {
	if nb.electionTrigger == nil {
		nb.electionTrigger = electionTrigger
	}
	return nb
}

func (nb *NodeBuilder) WithPK(publicKey types.PublicKey) *NodeBuilder {
	if nb.publicKey == "" {
		nb.publicKey = publicKey
	}
	return nb
}

func (nb *NodeBuilder) buildConfig() *leanhelix.Config {
	return &leanhelix.Config{
		ElectionTrigger: nb.electionTrigger,
	}
}

func (nb *NodeBuilder) Build() *network.Node {
	return network.NewNode(nb.publicKey, nb.buildConfig())
}

package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
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
	if nb.publicKey.Equals(lh.PublicKey("")) {
		nb.publicKey = publicKey
	}
	return nb
}

func (nb *NodeBuilder) buildConfig() lh.Config {
	return &mockConfig{
		electionTrigger: nb.electionTrigger,
	}
}

func (nb *NodeBuilder) Build() *Node {
	return NewNode(nb.publicKey, nb.buildConfig())
}

package builders

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/gossip"
)

type NodeBuilder struct {
	gossip     *gossip.Gossip
	blocksPool []lh.Block
	publicKey  primitives.Ed25519PublicKey
}

func NewNodeBuilder() *NodeBuilder {
	return &NodeBuilder{}
}

func (builder *NodeBuilder) ThatIsPartOf(gossip *gossip.Gossip) *NodeBuilder {
	if builder.gossip == nil {
		builder.gossip = gossip
	}
	return builder
}

func (builder *NodeBuilder) WithPublicKey(publicKey primitives.Ed25519PublicKey) *NodeBuilder {
	if builder.publicKey == nil {
		builder.publicKey = publicKey
	}
	return builder
}

func (builder *NodeBuilder) WithBlocksPool(blocksPool []lh.Block) *NodeBuilder {
	if builder.blocksPool == nil {
		builder.blocksPool = blocksPool
	}
	return builder
}

func (builder *NodeBuilder) Build() *Node {
	publicKey := builder.publicKey
	if publicKey == nil {
		publicKey = primitives.Ed25519PublicKey(fmt.Sprintf("Dummy PublicKey"))
	}

	blockUtils := NewMockBlockUtils(builder.blocksPool)
	electionTrigger := NewMockElectionTrigger()
	return NewNode(publicKey, builder.gossip, blockUtils, electionTrigger)
}

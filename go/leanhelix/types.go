package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/go/block"
)

type MessageType string

const (
	MESSAGE_TYPE_PREPREPARE  MessageType = "preprepare"
	MESSAGE_TYPE_PREPARE     MessageType = "prepare"
	MESSAGE_TYPE_COMMIT      MessageType = "commit"
	MESSAGE_TYPE_VIEW_CHANGE MessageType = "view-change"
	MESSAGE_TYPE_NEW_VIEW    MessageType = "new-view"
)

type Node struct {
	PublicKey []byte
}

type BlockUtils interface {
	CalculateBlockHash(block *block.Block) []byte
}

type NetworkCommunication interface {
	Nodes() []Node
	SendToMembers(publicKeys []string, messageType string, message []byte)
}

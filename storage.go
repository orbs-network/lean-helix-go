package leanhelix

import (
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type Storage interface {
	StorePreprepare(ppm *PreprepareMessage) bool
	StorePrepare(pp *PrepareMessage) bool
	StoreCommit(cm *CommitMessage) bool
	StoreViewChange(vcm *ViewChangeMessage) bool
	GetPrepareSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519PublicKey
	GetCommitSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519PublicKey
	GetViewChangeMessages(blockHeight BlockHeight, view View, f int) []*ViewChangeMessage
	GetPreprepare(blockHeight BlockHeight, view View) (*PreprepareMessage, bool)
	GetPreprepareBlock(blockHeight BlockHeight, view View) (Block, bool)
	GetPrepares(blockHeight BlockHeight, view View, blockHash Uint256) ([]*PrepareMessage, bool)
	GetLatestPreprepare(blockHeight BlockHeight) (*PreprepareMessage, bool)
	ClearTermLogs(blockHeight BlockHeight)
}

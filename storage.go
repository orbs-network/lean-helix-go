package leanhelix

import (
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type Storage interface {
	StorePreprepare(ppm *PreprepareMessage) bool
	GetPreprepareMessage(blockHeight BlockHeight, view View) (*PreprepareMessage, bool)
	GetPreprepareBlock(blockHeight BlockHeight, view View) (Block, bool)
	GetLatestPreprepare(blockHeight BlockHeight) (*PreprepareMessage, bool)

	StorePrepare(pp *PrepareMessage) bool
	GetPrepareMessages(blockHeight BlockHeight, view View, blockHash Uint256) ([]*PrepareMessage, bool)
	GetPrepareSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519PublicKey

	StoreCommit(cm *CommitMessage) bool
	GetCommitMessages(blockHeight BlockHeight, view View, blockHash Uint256) ([]*CommitMessage, bool)
	GetCommitSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519PublicKey

	StoreViewChange(vcm *ViewChangeMessage) bool
	GetViewChangeMessages(blockHeight BlockHeight, view View) ([]*ViewChangeMessage, bool)

	ClearBlockHeightLogs(blockHeight BlockHeight)
}

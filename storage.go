package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type Storage interface {
	StorePreprepare(ppm *PreprepareMessage) bool
	GetPreprepareMessage(blockHeight primitives.BlockHeight, view primitives.View) (*PreprepareMessage, bool)
	GetPreprepareBlock(blockHeight primitives.BlockHeight, view primitives.View) (Block, bool)
	GetLatestPreprepare(blockHeight primitives.BlockHeight) (*PreprepareMessage, bool)

	StorePrepare(pp *PrepareMessage) bool
	GetPrepareMessages(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) ([]*PrepareMessage, bool)
	GetPrepareSendersIds(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) []primitives.MemberId

	StoreCommit(cm *CommitMessage) bool
	GetCommitMessages(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) ([]*CommitMessage, bool)
	GetCommitSendersIds(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) []primitives.MemberId

	StoreViewChange(vcm *ViewChangeMessage) bool
	GetViewChangeMessages(blockHeight primitives.BlockHeight, view primitives.View) ([]*ViewChangeMessage, bool)

	ClearBlockHeightLogs(blockHeight primitives.BlockHeight)
}

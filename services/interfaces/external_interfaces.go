package interfaces

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/metrics"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"time"
)

type OnCommitCallback func(ctx context.Context, block Block, blockProof []byte)

type Config struct {
	InstanceId      primitives.InstanceId
	Communication   Communication
	Membership      Membership
	BlockUtils      BlockUtils
	KeyManager      KeyManager
	ElectionTrigger ElectionTrigger // TimerBasedElectionTrigger can be used
	Storage         Storage         // optional
	Logger          Logger          // optional
}

type ConsensusRawMessage struct {
	Content []byte
	Block   Block
}

type Communication interface {
	SendConsensusMessage(ctx context.Context, recipients []primitives.MemberId, message *ConsensusRawMessage)
}

type Membership interface {
	MyMemberId() primitives.MemberId
	RequestOrderedCommittee(ctx context.Context, blockHeight primitives.BlockHeight, randomSeed uint64) ([]primitives.MemberId, error)
}

type BlockUtils interface {
	RequestNewBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, prevBlock Block) (Block, primitives.BlockHash)
	ValidateBlockProposal(ctx context.Context, blockHeight primitives.BlockHeight, block Block, blockHash primitives.BlockHash, prevBlock Block) error
	ValidateBlockCommitment(blockHeight primitives.BlockHeight, block Block, blockHash primitives.BlockHash) bool
}

type KeyManager interface {
	SignConsensusMessage(blockHeight primitives.BlockHeight, content []byte) primitives.Signature
	VerifyConsensusMessage(blockHeight primitives.BlockHeight, content []byte, sender *protocol.SenderSignature) error
	SignRandomSeed(blockHeight primitives.BlockHeight, content []byte) primitives.RandomSeedSignature
	VerifyRandomSeed(blockHeight primitives.BlockHeight, content []byte, sender *protocol.SenderSignature) error
	AggregateRandomSeed(blockHeight primitives.BlockHeight, randomSeedShares []*protocol.SenderSignature) primitives.RandomSeedSignature
}

type ElectionTrigger interface {
	RegisterOnElection(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, cb func(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics)))
	Stop()
	ElectionChannel() chan func(ctx context.Context)
	CalcTimeout(view primitives.View) time.Duration
}

type Storage interface {
	StorePreprepare(ppm *PreprepareMessage) bool
	GetPreprepareMessage(blockHeight primitives.BlockHeight, view primitives.View) (*PreprepareMessage, bool)
	GetPreprepareBlock(blockHeight primitives.BlockHeight, view primitives.View) (Block, bool)
	GetLatestPreprepare(blockHeight primitives.BlockHeight) (*PreprepareMessage, bool)
	GetPreprepareFromView(blockHeight primitives.BlockHeight, view primitives.View) (*PreprepareMessage, bool)

	StorePrepare(pp *PrepareMessage) bool
	GetPrepareMessages(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) ([]*PrepareMessage, bool)
	GetPrepareSendersIds(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) []primitives.MemberId

	StoreCommit(cm *CommitMessage) bool
	GetCommitMessages(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) ([]*CommitMessage, bool)

	StoreViewChange(vcm *ViewChangeMessage) bool
	GetViewChangeMessages(blockHeight primitives.BlockHeight, view primitives.View) ([]*ViewChangeMessage, bool)

	ClearBlockHeightLogs(blockHeight primitives.BlockHeight)
}

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
}

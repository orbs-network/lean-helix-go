// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package interfaces

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/metrics"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"time"
)

type OnCommitCallback func(ctx context.Context, block Block, blockProof []byte)
type OnNewConsensusRoundCallback func(ctx context.Context, prevBlock Block, canBeFirstLeader bool)
type OnUpdateStateCallback func(ctx context.Context, currentHeight primitives.BlockHeight, receivedBlockHeight primitives.BlockHeight)

type Config struct {
	InstanceId              primitives.InstanceId
	Communication           Communication
	Membership              Membership
	BlockUtils              BlockUtils
	KeyManager              KeyManager
	ElectionTimeoutOnV0     time.Duration
	OnElectionCB            func(metrics.ElectionMetrics)
	Storage                 Storage // optional
	Logger                  Logger  // optional
	MsgChanBufLen           uint64
	UpdateStateChanBufLen   uint64
	ElectionChanBufLen      uint64
	OverrideElectionTrigger ElectionTrigger
}

type ConsensusRawMessage struct {
	Content []byte
	Block   Block
}

type Communication interface {
	SendConsensusMessage(ctx context.Context, recipients []primitives.MemberId, message *ConsensusRawMessage) error
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
	SignConsensusMessage(ctx context.Context, blockHeight primitives.BlockHeight, content []byte) primitives.Signature
	VerifyConsensusMessage(blockHeight primitives.BlockHeight, content []byte, sender *protocol.SenderSignature) error
	SignRandomSeed(ctx context.Context, blockHeight primitives.BlockHeight, content []byte) primitives.RandomSeedSignature
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

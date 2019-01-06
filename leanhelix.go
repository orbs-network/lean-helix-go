package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/blockheight"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/leanhelixterm"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/proofsvalidator"
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/orbs-network/lean-helix-go/services/rawmessagesfilter"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

type LeanHelix struct {
	messagesChannel         chan *interfaces.ConsensusRawMessage
	acknowledgeBlockChannel chan interfaces.Block
	currentHeight           primitives.BlockHeight
	config                  *interfaces.Config
	logger                  interfaces.Logger
	filter                  *rawmessagesfilter.RawMessageFilter
	leanHelixTerm           *leanhelixterm.LeanHelixTerm
	onCommitCallback        interfaces.OnCommitCallback
}

// ***********************************
// LeanHelix Constructor
// ***********************************
func NewLeanHelix(config *interfaces.Config, onCommitCallback interfaces.OnCommitCallback) *LeanHelix {
	if config.Logger == nil {
		config.Logger = logger.NewSilentLogger()
	}

	config.Logger.Debug("NewLeanHelix() ID=%s", termincommittee.Str(config.Membership.MyMemberId()))
	filter := rawmessagesfilter.NewConsensusMessageFilter(config.Membership.MyMemberId(), config.Logger)
	return &LeanHelix{
		messagesChannel:         make(chan *interfaces.ConsensusRawMessage),
		acknowledgeBlockChannel: make(chan interfaces.Block),
		currentHeight:           0,
		config:                  config,
		logger:                  config.Logger,
		filter:                  filter,
		onCommitCallback:        onCommitCallback,
	}
}

func (lh *LeanHelix) Run(ctx context.Context) {
	lh.logger.Info("Run() starting infinite loop")
	for {
		if !lh.Tick(ctx) {
			lh.logger.Info("Run() stopped infinite loop")
			return
		}
	}
}

func (lh *LeanHelix) UpdateState(ctx context.Context, prevBlock interfaces.Block, blockProofBytes []byte) {
	var height primitives.BlockHeight
	if prevBlock == nil {
		height = 0
	} else {
		height = prevBlock.Height()
	}
	lh.logger.Debug("LHFLOW UpdateState() ID=%s prevBlockHeight=%d", termincommittee.Str(lh.config.Membership.MyMemberId()), height)

	select {
	case <-ctx.Done():
		return

	case lh.acknowledgeBlockChannel <- prevBlock:
	}

}

//func (lh *LeanHelix) ValidateBlockConsensusGad(ctx context.Context, block interfaces.Block, blockProofBytes []byte, prevBlockProofBytes []byte) bool {
//	blockProof := protocol.BlockProofReader(blockProofBytes)
//	prevBlockProof := protocol.BlockProofReader(prevBlockProofBytes)
//	blockRef := blockProof.BlockRef()
//	blockHeight := blockRef.BlockHeight()
//	senderSignaturesIterator := blockProof.NodesIterator()
//	randomSeedSignature := blockProof.RandomSeedSignature()
//	prevRandomSeedSignature := prevBlockProof.RandomSeedSignature()
//
//	// Calculate the random seed based on prev block proof
//	randomSeed := calculateRandomSeed(prevRandomSeedSignature)
//	// validate random seed signature against master publicKey
//	randomSeedBytes := randomSeedToBytes(randomSeed)
//
//	masterRandomSeed := (&protocol.SenderSignatureBuilder{ // MemberId = 0\empty => master-group-id
//		Signature: primitives.Signature(randomSeedSignature),
//	}).Build()
//	if !lh.config.KeyManager.VerifyRandomSeed(blockHeight, randomSeedBytes, masterRandomSeed) {
//		return false
//	}
//
//	committeeMembers := lh.config.Membership.RequestOrderedCommittee(ctx, blockHeight, randomSeed, lh.config.maxCommitteeSize)
//	if !ValidatePBFTProof(blockRef, senderSignaturesIterator, quorumSize(committeeMembers), lh.config.KeyManager, committeeMembers) {
//		lh.config.Logger.Debug("ValidateBlockConsensus(): failed ValidateCommitProof()")
//		return false
//	}
//
//	return true
//}
//
//func calculateRandomSeed(signature []byte) uint64 {
//	hash := sha256.Sum256(signature)
//	randomSeed := uint64(0)
//	buf := bytes.NewBuffer(hash[:])
//	err := binary.Read(buf, binary.LittleEndian, &randomSeed)
//	if err != nil {
//		log.Fatalf("calculateRandomSeed decode failed: %s", err)
//	}
//	return randomSeed
//}
//
//func randomSeedToBytes(randomSeed uint64) []byte {
//	randomSeedBytes := make([]byte, 8)
//	binary.LittleEndian.PutUint64(randomSeedBytes, uint64(randomSeed))
//	return randomSeedBytes
//}

func ValidatePBFTProof(
	blockRef *protocol.BlockRef,
	cSendersIterator *protocol.BlockProofNodesIterator,
	q int,
	keyManager interfaces.KeyManager,
	membersIds []primitives.MemberId) bool {

	commitCount := 0

	for {
		if !cSendersIterator.HasNext() {
			break
		}
		cSender := cSendersIterator.NextNodes()
		cSenderMemberId := cSender.MemberId()
		if proofsvalidator.IsInMembers(membersIds, cSenderMemberId) == false {
			return false
		}
		if !proofsvalidator.VerifyBlockRefMessage(blockRef, cSender, keyManager) {
			return false
		}
		commitCount++
	}

	if commitCount < q {
		return false
	}

	return true
}

func (lh *LeanHelix) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProofBytes []byte, prevBlockProofBytes []byte) bool {
	lh.logger.Debug("ValidateBlockConsensus() ID=%s", termincommittee.Str(lh.config.Membership.MyMemberId()))
	if blockProofBytes == nil || len(blockProofBytes) == 0 || block == nil {
		return false
	}

	blockProof := protocol.BlockProofReader(blockProofBytes)
	blockRef := blockProof.BlockRef()
	if blockRef.MessageType() != protocol.LEAN_HELIX_COMMIT {
		return false
	}

	blockHeight := block.Height()
	if blockHeight != blockRef.BlockHeight() {
		return false
	}

	if !lh.config.BlockUtils.ValidateBlockCommitment(blockHeight, block, blockRef.BlockHash()) {
		return false
	}

	committeeMembers := lh.config.Membership.RequestOrderedCommittee(ctx, blockHeight, 0)

	sendersIterator := blockProof.NodesIterator()
	set := make(map[storage.MemberIdStr]bool)
	var sendersCounter = 0
	for {
		if !sendersIterator.HasNext() {
			break
		}

		sender := sendersIterator.NextNodes()
		if !proofsvalidator.VerifyBlockRefMessage(blockRef, sender, lh.config.KeyManager) {
			return false
		}

		memberId := sender.MemberId()
		if _, ok := set[storage.MemberIdStr(memberId)]; ok {
			return false
		}

		if !proofsvalidator.IsInMembers(committeeMembers, memberId) {
			return false
		}

		set[storage.MemberIdStr(memberId)] = true
		sendersCounter++
	}

	if sendersCounter < quorum.CalcQuorumSize(len(committeeMembers)) {
		return false
	}

	if len(blockProof.RandomSeedSignature()) == 0 || blockProof.RandomSeedSignature() == nil {
		return false
	}

	return true
}

func (lh *LeanHelix) HandleConsensusMessage(ctx context.Context, message *interfaces.ConsensusRawMessage) {
	lh.logger.Debug("HandleConsensusRawMessage() ID=%s", termincommittee.Str(lh.config.Membership.MyMemberId()))
	select {
	case <-ctx.Done():
		return

	case lh.messagesChannel <- message:
	}
}

func (lh *LeanHelix) Tick(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		lh.logger.Debug("LHFLOW Tick Done")
		return false

	case message := <-lh.messagesChannel:
		//lh.logger.Debug("LHFLOW Tick Message")
		lh.filter.HandleConsensusRawMessage(ctx, message)

	case trigger := <-lh.config.ElectionTrigger.ElectionChannel():
		lh.logger.Debug("LHFLOW Tick Election")
		if trigger == nil {
			lh.logger.Debug("LHFLOW Tick Election, OMG trigger is nil!")
		}
		trigger(ctx)

	case prevBlock := <-lh.acknowledgeBlockChannel:
		lh.logger.Debug("LHFLOW Tick Update")
		// TODO: a byzantine node can send the genesis block in sync can cause a mess
		prevHeight := blockheight.GetBlockHeight(prevBlock)
		if prevHeight >= lh.currentHeight {
			lh.logger.Debug("Calling onNewConsensusRound() from Tick() prevHeight=%d lh.currentHeight=%d", prevHeight, lh.currentHeight)
			lh.onNewConsensusRound(ctx, prevBlock)
		}
	}

	return true
}

// ************************ Internal ***************************************

func (lh *LeanHelix) onCommit(ctx context.Context, block interfaces.Block, blockProof []byte) {
	lh.onCommitCallback(ctx, block, blockProof)
	lh.logger.Debug("Calling onNewConsensusRound() from onCommit() lh.currentHeight=%d", lh.currentHeight)
	lh.onNewConsensusRound(ctx, block)
}

func (lh *LeanHelix) onNewConsensusRound(ctx context.Context, prevBlock interfaces.Block) {
	lh.currentHeight = blockheight.GetBlockHeight(prevBlock) + 1
	lh.leanHelixTerm = leanhelixterm.NewLeanHelixTerm(ctx, lh.config, lh.onCommit, prevBlock)
	lh.filter.SetBlockHeight(ctx, lh.currentHeight, lh.leanHelixTerm)
}

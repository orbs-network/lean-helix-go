package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"math"
	"sort"
)

// The algorithm cannot function with less committee members
// because it cannot calculate the f number (where committee members are 3f+1)
// The only reason to set this manually in config below this limit is for internal tests
const LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS = 4

type TermInCommittee struct {
	keyManager                     KeyManager
	communication                  Communication
	storage                        Storage
	electionTrigger                ElectionTrigger
	blockUtils                     BlockUtils
	onCommit                       OnCommitCallback
	messageFactory                 *MessageFactory
	myMemberId                     primitives.MemberId
	committeeMembersMemberIds      []primitives.MemberId
	otherCommitteeMembersMemberIds []primitives.MemberId
	height                         primitives.BlockHeight
	view                           primitives.View
	preparedLocally                bool
	committedBlock                 Block
	leaderMemberId                 primitives.MemberId
	newViewLocally                 primitives.View
	logger                         Logger
	prevBlock                      Block
}

func NewTermInCommittee(ctx context.Context, config *Config, onCommit OnCommitCallback, prevBlock Block) *TermInCommittee {
	keyManager := config.KeyManager
	blockUtils := config.BlockUtils
	membership := config.Membership
	myMemberId := membership.MyMemberId()
	comm := config.Communication
	messageFactory := NewMessageFactory(keyManager, myMemberId)

	var prevBlockHeight primitives.BlockHeight
	if prevBlock == GenesisBlock {
		prevBlockHeight = 0
	} else {
		prevBlockHeight = prevBlock.Height()
	}
	newBlockHeight := prevBlockHeight + 1

	// TODO Implement me!
	randomSeed := uint64(12345)
	// TODO Implement me!
	committeeSize := uint32(4)
	committeeMembers := membership.RequestOrderedCommittee(ctx, newBlockHeight, randomSeed, committeeSize)

	panicOnLessThanMinimumCommitteeMembers(config.OverrideMinimumCommitteeMembers, committeeMembers)

	otherCommitteeMembers := make([]primitives.MemberId, 0)
	for _, member := range committeeMembers {
		if !member.Equal(myMemberId) {
			otherCommitteeMembers = append(otherCommitteeMembers, member)
		}
	}
	if config.Logger == nil {
		config.Logger = NewSilentLogger()
	}

	if config.Storage == nil {
		config.Storage = NewInMemoryStorage()
	}

	newTerm := &TermInCommittee{
		onCommit:                       onCommit,
		height:                         newBlockHeight,
		prevBlock:                      prevBlock,
		keyManager:                     keyManager,
		communication:                  comm,
		storage:                        config.Storage,
		electionTrigger:                config.ElectionTrigger,
		blockUtils:                     blockUtils,
		committeeMembersMemberIds:      committeeMembers,
		otherCommitteeMembersMemberIds: otherCommitteeMembers,
		messageFactory:                 messageFactory,
		myMemberId:                     myMemberId,
		logger:                         config.Logger,
	}

	newTerm.logger.Debug("H=%d V=0 %s NewTermInCommittee: committeeMembersCount=%d", newBlockHeight, myMemberId.KeyForMap(), len(committeeMembers))
	newTerm.initView(0)
	return newTerm
}

func panicOnLessThanMinimumCommitteeMembers(minimum int, committeeMembers []primitives.MemberId) {

	if minimum == 0 {
		minimum = LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS
	}

	if len(committeeMembers) < minimum {
		panic(fmt.Sprintf("LH Received only %d committee members, but the hard minimum is %d", len(committeeMembers), LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS))
	}
}

func (tic *TermInCommittee) StartTerm(ctx context.Context) {
	if tic.IsLeader() {
		block, blockHash := tic.blockUtils.RequestNewBlockProposal(ctx, tic.height, tic.prevBlock)
		ppm := tic.messageFactory.CreatePreprepareMessage(tic.height, tic.view, block, blockHash)

		tic.storage.StorePreprepare(ppm)
		tic.sendConsensusMessage(ctx, ppm)
	}
}

func (tic *TermInCommittee) GetView() primitives.View {
	return tic.view
}

func (tic *TermInCommittee) SetView(view primitives.View) {
	if tic.view != view {
		tic.initView(view)
	}
}

func (tic *TermInCommittee) initView(view primitives.View) {
	tic.preparedLocally = false
	tic.view = view
	tic.leaderMemberId = tic.calcLeaderMemberId(view)
	tic.electionTrigger.RegisterOnElection(tic.height, tic.view, tic.moveToNextLeader)
	tic.logger.Debug("H=%d V=%d %s initView() set leader to %s", tic.height, tic.view, tic.myMemberId.KeyForMap(), tic.leaderMemberId.KeyForMap())
}

func (tic *TermInCommittee) Dispose() {
	tic.storage.ClearBlockHeightLogs(tic.height)
}

func (tic *TermInCommittee) calcLeaderMemberId(view primitives.View) primitives.MemberId {
	index := int(view) % len(tic.committeeMembersMemberIds)
	return tic.committeeMembersMemberIds[index]
}

func (tic *TermInCommittee) moveToNextLeader(ctx context.Context, height primitives.BlockHeight, view primitives.View) {
	if view != tic.view || height != tic.height {
		return
	}
	tic.SetView(tic.view + 1)
	tic.logger.Debug("H=%d V=%d moveToNextLeader() newLeader=%s", tic.height, tic.view, tic.leaderMemberId[:3])
	preparedMessages := ExtractPreparedMessages(tic.height, tic.storage, tic.QuorumSize())
	vcm := tic.messageFactory.CreateViewChangeMessage(tic.height, tic.view, preparedMessages)
	if tic.IsLeader() {
		tic.storage.StoreViewChange(vcm)
		tic.checkElected(ctx, tic.height, tic.view)
	} else {
		tic.sendConsensusMessage(ctx, vcm)
	}
}

func (tic *TermInCommittee) sendConsensusMessage(ctx context.Context, message ConsensusMessage) {
	tic.logger.Debug("H=%d V=%d %s sendConsensusMessage() msgType=%v", tic.height, tic.view, tic.myMemberId.KeyForMap(), message.MessageType())
	rawMessage := CreateConsensusRawMessage(message)
	tic.communication.SendConsensusMessage(ctx, tic.otherCommitteeMembersMemberIds, rawMessage)
}

func (tic *TermInCommittee) HandleLeanHelixPrePrepare(ctx context.Context, ppm *PreprepareMessage) {
	tic.logger.Debug("H=%s V=%s %s HandleLeanHelixPrePrepare()", tic.height, tic.view, tic.myMemberId.KeyForMap())
	if err := tic.validatePreprepare(ctx, ppm); err != nil {
		tic.logger.Debug("H=%s V=%s HandleLeanHelixPrePrepare() err=%v", err)
	} else {
		tic.processPreprepare(ctx, ppm)
	}
}

func (tic *TermInCommittee) processPreprepare(ctx context.Context, ppm *PreprepareMessage) {
	header := ppm.content.SignedHeader()
	if tic.view != header.View() {
		tic.logger.Debug("H=%s V=%s processPreprepare() message from incorrect view %d", tic.height, tic.view, header.View())
		return
	}

	pm := tic.messageFactory.CreatePrepareMessage(header.BlockHeight(), header.View(), header.BlockHash())
	tic.storage.StorePreprepare(ppm)
	tic.storage.StorePrepare(pm)
	tic.sendConsensusMessage(ctx, pm)
	tic.checkPrepared(ctx, header.BlockHeight(), header.View(), header.BlockHash())
}

func (tic *TermInCommittee) validatePreprepare(ctx context.Context, ppm *PreprepareMessage) error {
	blockHeight := ppm.BlockHeight()
	view := ppm.View()
	if tic.hasPreprepare(blockHeight, view) {
		return fmt.Errorf("already received Preprepare for H=%s V=%s", blockHeight, view)
	}

	header := ppm.Content().SignedHeader()
	sender := ppm.Content().Sender()
	if !tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		return fmt.Errorf("verification failed for sender %s signature on header", sender.MemberId()[:3])
	}

	leaderMemberId := tic.calcLeaderMemberId(view)
	senderMemberId := sender.MemberId()
	if !senderMemberId.Equal(leaderMemberId) {
		// Log
		return fmt.Errorf("sender %s is not leader", senderMemberId[:3])
	}

	isValidBlock := tic.blockUtils.ValidateBlockProposal(ctx, blockHeight, ppm.Block(), ppm.Content().SignedHeader().BlockHash(), tic.prevBlock)

	if !isValidBlock {
		return fmt.Errorf("block validation failed")
	}

	return nil
}

func (tic *TermInCommittee) hasPreprepare(blockHeight primitives.BlockHeight, view primitives.View) bool {
	_, ok := tic.storage.GetPreprepareMessage(blockHeight, view)
	return ok
}

func (tic *TermInCommittee) HandleLeanHelixPrepare(ctx context.Context, pm *PrepareMessage) {
	tic.logger.Debug("H=%s V=%s %s HandleLeanHelixPrepare()", pm.BlockHeight(), pm.View(), tic.myMemberId.KeyForMap())
	header := pm.content.SignedHeader()
	sender := pm.content.Sender()

	if !tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		tic.logger.Info("verification failed for Prepare blockHeight=%v view=%v blockHash=%v", header.BlockHeight(), header.View(), header.BlockHash())
		return
	}
	if tic.view > header.View() {
		tic.logger.Info("prepare view %v is less than OneHeight's view %v", header.View(), tic.view)
		return
	}
	if tic.leaderMemberId.Equal(sender.MemberId()) {
		tic.logger.Info("prepare received from leader (only preprepare can be received from leader)")
		return
	}
	tic.storage.StorePrepare(pm)
	if tic.view == header.View() {
		tic.checkPrepared(ctx, header.BlockHeight(), header.View(), header.BlockHash())
	}
}

func (tic *TermInCommittee) HandleLeanHelixViewChange(ctx context.Context, vcm *ViewChangeMessage) {
	tic.logger.Debug("H=%s V=%s HandleLeanHelixViewChange()", tic.height, tic.view)
	if !tic.isViewChangeValid(tic.myMemberId, tic.view, vcm.content) {
		tic.logger.Info("message ViewChange is not valid")
		return
	}

	header := vcm.content.SignedHeader()
	if vcm.block != nil && header.PreparedProof() != nil {
		isValidDigest := tic.blockUtils.ValidateBlockCommitment(vcm.BlockHeight(), vcm.block, header.PreparedProof().PreprepareBlockRef().BlockHash())
		if !isValidDigest {
			tic.logger.Info("different block hashes for block provided with message, and the block provided by the PPM in the PreparedProof of the message")
			return
		}
	}

	tic.storage.StoreViewChange(vcm)
	tic.checkElected(ctx, header.BlockHeight(), header.View())
}

func (tic *TermInCommittee) isViewChangeValid(targetLeaderMemberId primitives.MemberId, view primitives.View, confirmation *protocol.ViewChangeMessageContent) bool {
	header := confirmation.SignedHeader()
	sender := confirmation.Sender()
	newView := header.View()
	preparedProof := header.PreparedProof()

	if !tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		tic.logger.Debug("isViewChangeValid(): VerifyConsensusMessage() failed")
		isVerified := tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender)
		tic.logger.Debug("isViewChangeValid(): isVerified %t", isVerified)
		return false
	}

	if view > newView {
		tic.logger.Debug("isViewChangeValid(): message from old view")
		return false
	}

	if !ValidatePreparedProof(tic.height, newView, preparedProof, tic.QuorumSize(), tic.keyManager, tic.committeeMembersMemberIds, func(view primitives.View) primitives.MemberId { return tic.calcLeaderMemberId(view) }) {
		tic.logger.Debug("isViewChangeValid(): failed ValidatePreparedProof()")
		return false
	}

	futureLeaderMemberId := tic.calcLeaderMemberId(newView)
	if !targetLeaderMemberId.Equal(futureLeaderMemberId) {
		return false
	}

	return true

}

func (tic *TermInCommittee) checkElected(ctx context.Context, height primitives.BlockHeight, view primitives.View) {
	if tic.newViewLocally < view {
		vcms, ok := tic.storage.GetViewChangeMessages(height, view)
		minimumNodes := tic.QuorumSize()
		if ok && len(vcms) >= minimumNodes {
			tic.onElected(ctx, view, vcms[:minimumNodes])
		}
	}
}

func (tic *TermInCommittee) onElected(ctx context.Context, view primitives.View, viewChangeMessages []*ViewChangeMessage) {
	tic.newViewLocally = view
	tic.SetView(view)
	block := GetLatestBlockFromViewChangeMessages(viewChangeMessages)
	var blockHash primitives.BlockHash
	if block == nil {
		block, blockHash = tic.blockUtils.RequestNewBlockProposal(ctx, tic.height, tic.prevBlock)
	}
	ppmContentBuilder := tic.messageFactory.CreatePreprepareMessageContentBuilder(tic.height, view, block, blockHash)
	ppm := tic.messageFactory.CreatePreprepareMessageFromContentBuilder(ppmContentBuilder, block)
	confirmations := extractConfirmationsFromViewChangeMessages(viewChangeMessages)
	nvm := tic.messageFactory.CreateNewViewMessage(tic.height, view, ppmContentBuilder, confirmations, block)
	tic.storage.StorePreprepare(ppm)
	tic.sendConsensusMessage(ctx, nvm)
}

func (tic *TermInCommittee) checkPrepared(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	if tic.preparedLocally == false {
		if tic.isPreprepared(blockHeight, view, blockHash) {
			countPrepared := tic.countPrepared(blockHeight, view, blockHash)
			if countPrepared >= tic.QuorumSize()-1 {
				tic.onPrepared(ctx, blockHeight, view, blockHash)
			}
		}
	}
}

func (tic *TermInCommittee) onPrepared(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	tic.preparedLocally = true
	cm := tic.messageFactory.CreateCommitMessage(blockHeight, view, blockHash)
	tic.storage.StoreCommit(cm)
	tic.sendConsensusMessage(ctx, cm)
	tic.checkCommitted(ctx, blockHeight, view, blockHash)
}

func (tic *TermInCommittee) HandleLeanHelixCommit(ctx context.Context, cm *CommitMessage) {
	tic.logger.Debug("H=%s V=%s %s HandleLeanHelixCommit()", tic.height, tic.view, tic.myMemberId.KeyForMap())
	header := cm.content.SignedHeader()
	sender := cm.content.Sender()

	if !tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		tic.logger.Debug("verification failed for Commit blockHeight=%v view=%v blockHash=%v", header.BlockHeight(), header.View(), header.BlockHash())
		return
	}
	tic.storage.StoreCommit(cm)
	tic.checkCommitted(ctx, header.BlockHeight(), header.View(), header.BlockHash())
}

func (tic *TermInCommittee) checkCommitted(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	tic.logger.Debug("H=%s V=%s %s checkCommitted() H=%s V=%s BlockHash %s ", tic.height, tic.view, tic.myMemberId.KeyForMap(), blockHeight, view, blockHash)
	if tic.committedBlock != nil {
		return
	}
	if !tic.isPreprepared(blockHeight, view, blockHash) {
		return
	}
	commits := tic.storage.GetCommitSendersIds(blockHeight, view, blockHash)
	if len(commits) < tic.QuorumSize() {
		return
	}
	ppm, ok := tic.storage.GetPreprepareMessage(blockHeight, view)
	if !ok {
		// log
		tic.logger.Info("H=%s V=%s checkCommitted() missing PPM")
		return
	}
	tic.logger.Info("H=%s V=%s %s checkCommitted() COMMITTED H=%s V=%s BlockHash %s ", tic.height, tic.view, tic.myMemberId.KeyForMap(), blockHeight, view, blockHash)
	tic.committedBlock = ppm.block
	tic.onCommit(ctx, ppm.block, nil)
}

func (tic *TermInCommittee) validateViewChangeVotes(targetBlockHeight primitives.BlockHeight, targetView primitives.View, confirmations []*protocol.ViewChangeMessageContent) bool {
	if len(confirmations) < tic.QuorumSize() {
		return false
	}

	set := make(map[string]bool)

	// VerifyConsensusMessage that all _Block heights and views match, and all public keys are unique
	for _, confirmation := range confirmations {
		senderMemberIdStr := string(confirmation.Sender().MemberId())
		if confirmation.SignedHeader().BlockHeight() != targetBlockHeight {
			return false
		}
		if confirmation.SignedHeader().View() != targetView {
			return false
		}
		if set[senderMemberIdStr] {
			return false
		}
		set[senderMemberIdStr] = true
	}

	return true

}

func (tic *TermInCommittee) latestViewChangeVote(confirmations []*protocol.ViewChangeMessageContent) *protocol.ViewChangeMessageContent {
	res := make([]*protocol.ViewChangeMessageContent, 0, len(confirmations))
	for _, confirmation := range confirmations {
		if confirmation.SignedHeader().PreparedProof() != nil && len(confirmation.SignedHeader().PreparedProof().Raw()) > 0 {
			res = append(res, confirmation)
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[j].SignedHeader().PreparedProof().PreprepareBlockRef().View() < res[i].SignedHeader().PreparedProof().PreprepareBlockRef().View()
	})

	if len(res) > 0 {
		return res[0]
	} else {
		return nil
	}
}

func (tic *TermInCommittee) HandleLeanHelixNewView(ctx context.Context, nvm *NewViewMessage) {
	tic.logger.Debug("H=%s V=%s %s HandleLeanHelixNewView()", tic.height, tic.view, tic.myMemberId.KeyForMap())
	header := nvm.Content().SignedHeader()
	sender := nvm.Content().Sender()
	ppMessageContent := nvm.Content().Message()
	viewChangeConfirmationsIter := header.ViewChangeConfirmationsIterator()
	viewChangeConfirmations := make([]*protocol.ViewChangeMessageContent, 0, 1)
	for {
		if !viewChangeConfirmationsIter.HasNext() {
			break
		}
		viewChangeConfirmations = append(viewChangeConfirmations, viewChangeConfirmationsIter.NextViewChangeConfirmations())
	}

	if !tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", ignored because the signature verification failed` });
		tic.logger.Debug("HandleLeanHelixNewView(): verify failed")
		return
	}

	futureLeaderId := tic.calcLeaderMemberId(header.View())
	if !sender.MemberId().Equal(futureLeaderId) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", rejected because it match the new id (${view})` });
		tic.logger.Debug("HandleLeanHelixNewView(): no match for future leader")
		return
	}

	if !tic.validateViewChangeVotes(header.BlockHeight(), header.View(), viewChangeConfirmations) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", votes is invalid` });
		tic.logger.Debug("HandleLeanHelixNewView(): validateViewChangeVotes failed")
		return
	}

	if tic.view > header.View() {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", view is from the past` });
		tic.logger.Debug("HandleLeanHelixNewView(): current view is higher than message view")
		return
	}

	if !ppMessageContent.SignedHeader().View().Equal(header.View()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", view doesn't match PP.view` });
		tic.logger.Debug("HandleLeanHelixNewView(): NewView.view and NewView.Preprepare.view do not match")
		return
	}

	if !ppMessageContent.SignedHeader().BlockHeight().Equal(header.BlockHeight()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", blockHeight doesn't match PP.blockHeight` });
		tic.logger.Debug("HandleLeanHelixNewView(): NewView.BlockHeight and NewView.Preprepare.BlockHeight do not match")
		return
	}

	latestVote := tic.latestViewChangeVote(viewChangeConfirmations)
	if latestVote != nil {
		viewChangeMessageValid := tic.isViewChangeValid(futureLeaderId, header.View(), latestVote)
		if !viewChangeMessageValid {
			//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", view change votes are invalid` });
			tic.logger.Debug("HandleLeanHelixNewView(): NewView.ViewChangeConfirmation (with latest view) is invalid")
			return
		}

		// rewrite this mess
		latestVoteBlockHash := latestVote.SignedHeader().PreparedProof().PreprepareBlockRef().BlockHash()
		if latestVoteBlockHash != nil {
			isValidDigest := tic.blockUtils.ValidateBlockCommitment(header.BlockHeight(), nvm.Block(), latestVoteBlockHash)
			if !isValidDigest {
				//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", the given _Block (PP._Block) doesn't match the best _Block from the VCProof` });
				tic.logger.Debug("HandleLeanHelixNewView(): NewView.ViewChangeConfirmation (with latest view) is invalid")
				return
			}
		}
	}

	ppm := &PreprepareMessage{
		content: ppMessageContent,
		block:   nvm.Block(),
	}

	if err := tic.validatePreprepare(ctx, ppm); err == nil {
		tic.newViewLocally = header.View()
		tic.SetView(header.View())
		tic.processPreprepare(ctx, ppm)
	}
}

func (tic *TermInCommittee) QuorumSize() int {
	committeeMembersCount := len(tic.committeeMembersMemberIds)
	f := int(math.Floor(float64(committeeMembersCount-1) / 3))
	return committeeMembersCount - f
}

func (tic *TermInCommittee) IsLeader() bool {
	return tic.myMemberId.Equal(tic.leaderMemberId)
}

func (tic *TermInCommittee) countPrepared(height primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) int {
	return len(tic.storage.GetPrepareSendersIds(height, view, blockHash))
}

func (tic *TermInCommittee) isPreprepared(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) bool {
	ppm, ok := tic.storage.GetPreprepareMessage(blockHeight, view)
	if !ok {
		return false
	}
	ppmBlock := ppm.block
	if ppmBlock == nil {
		return false
	}

	ppmBlockHash := ppm.Content().SignedHeader().BlockHash()
	return ppmBlockHash.Equal(blockHash)
}

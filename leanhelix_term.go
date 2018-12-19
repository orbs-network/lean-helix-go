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

type LeanHelixTerm struct {
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

func NewLeanHelixTerm(ctx context.Context, config *Config, onCommit OnCommitCallback, prevBlock Block) *LeanHelixTerm {
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

	newTerm := &LeanHelixTerm{
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

	newTerm.logger.Debug("H=%d V=0 %s NewLeanHelixTerm: committeeMembersCount=%d", newBlockHeight, myMemberId.KeyForMap(), len(committeeMembers))
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

func (term *LeanHelixTerm) StartTerm(ctx context.Context) {
	if term.IsLeader() {
		block, blockHash := term.blockUtils.RequestNewBlockProposal(ctx, term.height, term.prevBlock)
		ppm := term.messageFactory.CreatePreprepareMessage(term.height, term.view, block, blockHash)

		term.storage.StorePreprepare(ppm)
		term.sendConsensusMessage(ctx, ppm)
	}
}

func (term *LeanHelixTerm) GetView() primitives.View {
	return term.view
}

func (term *LeanHelixTerm) SetView(view primitives.View) {
	if term.view != view {
		term.initView(view)
	}
}

func (term *LeanHelixTerm) initView(view primitives.View) {
	term.preparedLocally = false
	term.view = view
	term.leaderMemberId = term.calcLeaderMemberId(view)
	term.electionTrigger.RegisterOnElection(term.height, term.view, term.moveToNextLeader)
	term.logger.Debug("H=%d V=%d %s initView() set leader to %s", term.height, term.view, term.myMemberId.KeyForMap(), term.leaderMemberId.KeyForMap())
}

func (term *LeanHelixTerm) Dispose() {
	term.storage.ClearBlockHeightLogs(term.height)
}

func (term *LeanHelixTerm) calcLeaderMemberId(view primitives.View) primitives.MemberId {
	index := int(view) % len(term.committeeMembersMemberIds)
	return term.committeeMembersMemberIds[index]
}

func (term *LeanHelixTerm) moveToNextLeader(ctx context.Context, height primitives.BlockHeight, view primitives.View) {
	if view != term.view || height != term.height {
		return
	}
	term.SetView(term.view + 1)
	term.logger.Debug("H=%d V=%d moveToNextLeader() newLeader=%s", term.height, term.view, term.leaderMemberId[:3])
	preparedMessages := ExtractPreparedMessages(term.height, term.storage, term.QuorumSize())
	vcm := term.messageFactory.CreateViewChangeMessage(term.height, term.view, preparedMessages)
	if term.IsLeader() {
		term.storage.StoreViewChange(vcm)
		term.checkElected(ctx, term.height, term.view)
	} else {
		term.sendConsensusMessage(ctx, vcm)
	}
}

func (term *LeanHelixTerm) sendConsensusMessage(ctx context.Context, message ConsensusMessage) {
	term.logger.Debug("H=%d V=%d %s sendConsensusMessage() msgType=%v", term.height, term.view, term.myMemberId.KeyForMap(), message.MessageType())
	rawMessage := CreateConsensusRawMessage(message)
	term.communication.SendConsensusMessage(ctx, term.otherCommitteeMembersMemberIds, rawMessage)
}

func (term *LeanHelixTerm) HandleLeanHelixPrePrepare(ctx context.Context, ppm *PreprepareMessage) {
	term.logger.Debug("H=%s V=%s %s HandleLeanHelixPrePrepare()", term.height, term.view, term.myMemberId.KeyForMap())
	if err := term.validatePreprepare(ctx, ppm); err != nil {
		term.logger.Debug("H=%s V=%s HandleLeanHelixPrePrepare() err=%v", err)
	} else {
		term.processPreprepare(ctx, ppm)
	}
}

func (term *LeanHelixTerm) processPreprepare(ctx context.Context, ppm *PreprepareMessage) {
	header := ppm.content.SignedHeader()
	if term.view != header.View() {
		term.logger.Debug("H=%s V=%s processPreprepare() message from incorrect view %d", term.height, term.view, header.View())
		return
	}

	pm := term.messageFactory.CreatePrepareMessage(header.BlockHeight(), header.View(), header.BlockHash())
	term.storage.StorePreprepare(ppm)
	term.storage.StorePrepare(pm)
	term.sendConsensusMessage(ctx, pm)
	term.checkPrepared(ctx, header.BlockHeight(), header.View(), header.BlockHash())
}

func (term *LeanHelixTerm) validatePreprepare(ctx context.Context, ppm *PreprepareMessage) error {
	blockHeight := ppm.BlockHeight()
	view := ppm.View()
	if term.hasPreprepare(blockHeight, view) {
		return fmt.Errorf("already received Preprepare for H=%s V=%s", blockHeight, view)
	}

	header := ppm.Content().SignedHeader()
	sender := ppm.Content().Sender()
	if !term.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		return fmt.Errorf("verification failed for sender %s signature on header", sender.MemberId()[:3])
	}

	leaderMemberId := term.calcLeaderMemberId(view)
	senderMemberId := sender.MemberId()
	if !senderMemberId.Equal(leaderMemberId) {
		// Log
		return fmt.Errorf("sender %s is not leader", senderMemberId[:3])
	}

	isValidBlock := term.blockUtils.ValidateBlockProposal(ctx, blockHeight, ppm.Block(), ppm.Content().SignedHeader().BlockHash(), term.prevBlock)

	if !isValidBlock {
		return fmt.Errorf("block validation failed")
	}

	return nil
}

func (term *LeanHelixTerm) hasPreprepare(blockHeight primitives.BlockHeight, view primitives.View) bool {
	_, ok := term.storage.GetPreprepareMessage(blockHeight, view)
	return ok
}

func (term *LeanHelixTerm) HandleLeanHelixPrepare(ctx context.Context, pm *PrepareMessage) {
	term.logger.Debug("H=%s V=%s %s HandleLeanHelixPrepare()", pm.BlockHeight(), pm.View(), term.myMemberId.KeyForMap())
	header := pm.content.SignedHeader()
	sender := pm.content.Sender()

	if !term.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		term.logger.Info("verification failed for Prepare blockHeight=%v view=%v blockHash=%v", header.BlockHeight(), header.View(), header.BlockHash())
		return
	}
	if term.view > header.View() {
		term.logger.Info("prepare view %v is less than OneHeight's view %v", header.View(), term.view)
		return
	}
	if term.leaderMemberId.Equal(sender.MemberId()) {
		term.logger.Info("prepare received from leader (only preprepare can be received from leader)")
		return
	}
	term.storage.StorePrepare(pm)
	if term.view == header.View() {
		term.checkPrepared(ctx, header.BlockHeight(), header.View(), header.BlockHash())
	}
}

func (term *LeanHelixTerm) HandleLeanHelixViewChange(ctx context.Context, vcm *ViewChangeMessage) {
	term.logger.Debug("H=%s V=%s HandleLeanHelixViewChange()", term.height, term.view)
	if !term.isViewChangeValid(term.myMemberId, term.view, vcm.content) {
		term.logger.Info("message ViewChange is not valid")
		return
	}

	header := vcm.content.SignedHeader()
	if vcm.block != nil && header.PreparedProof() != nil {
		isValidDigest := term.blockUtils.ValidateBlockCommitment(vcm.BlockHeight(), vcm.block, header.PreparedProof().PreprepareBlockRef().BlockHash())
		if !isValidDigest {
			term.logger.Info("different block hashes for block provided with message, and the block provided by the PPM in the PreparedProof of the message")
			return
		}
	}

	term.storage.StoreViewChange(vcm)
	term.checkElected(ctx, header.BlockHeight(), header.View())
}

func (term *LeanHelixTerm) isViewChangeValid(targetLeaderMemberId primitives.MemberId, view primitives.View, confirmation *protocol.ViewChangeMessageContent) bool {
	header := confirmation.SignedHeader()
	sender := confirmation.Sender()
	newView := header.View()
	preparedProof := header.PreparedProof()

	if !term.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		term.logger.Debug("isViewChangeValid(): VerifyConsensusMessage() failed")
		isVerified := term.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender)
		term.logger.Debug("isViewChangeValid(): isVerified %t", isVerified)
		return false
	}

	if view > newView {
		term.logger.Debug("isViewChangeValid(): message from old view")
		return false
	}

	if !ValidatePreparedProof(term.height, newView, preparedProof, term.QuorumSize(), term.keyManager, term.committeeMembersMemberIds, func(view primitives.View) primitives.MemberId { return term.calcLeaderMemberId(view) }) {
		term.logger.Debug("isViewChangeValid(): failed ValidatePreparedProof()")
		return false
	}

	futureLeaderMemberId := term.calcLeaderMemberId(newView)
	if !targetLeaderMemberId.Equal(futureLeaderMemberId) {
		return false
	}

	return true

}

func (term *LeanHelixTerm) checkElected(ctx context.Context, height primitives.BlockHeight, view primitives.View) {
	if term.newViewLocally < view {
		vcms, ok := term.storage.GetViewChangeMessages(height, view)
		minimumNodes := term.QuorumSize()
		if ok && len(vcms) >= minimumNodes {
			term.onElected(ctx, view, vcms[:minimumNodes])
		}
	}
}

func (term *LeanHelixTerm) onElected(ctx context.Context, view primitives.View, viewChangeMessages []*ViewChangeMessage) {
	term.newViewLocally = view
	term.SetView(view)
	block := GetLatestBlockFromViewChangeMessages(viewChangeMessages)
	var blockHash primitives.BlockHash
	if block == nil {
		block, blockHash = term.blockUtils.RequestNewBlockProposal(ctx, term.height, term.prevBlock)
	}
	ppmContentBuilder := term.messageFactory.CreatePreprepareMessageContentBuilder(term.height, view, block, blockHash)
	ppm := term.messageFactory.CreatePreprepareMessageFromContentBuilder(ppmContentBuilder, block)
	confirmations := extractConfirmationsFromViewChangeMessages(viewChangeMessages)
	nvm := term.messageFactory.CreateNewViewMessage(term.height, view, ppmContentBuilder, confirmations, block)
	term.storage.StorePreprepare(ppm)
	term.sendConsensusMessage(ctx, nvm)
}

func (term *LeanHelixTerm) checkPrepared(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	if term.preparedLocally == false {
		if term.isPreprepared(blockHeight, view, blockHash) {
			countPrepared := term.countPrepared(blockHeight, view, blockHash)
			if countPrepared >= term.QuorumSize()-1 {
				term.onPrepared(ctx, blockHeight, view, blockHash)
			}
		}
	}
}

func (term *LeanHelixTerm) onPrepared(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	term.preparedLocally = true
	cm := term.messageFactory.CreateCommitMessage(blockHeight, view, blockHash)
	term.storage.StoreCommit(cm)
	term.sendConsensusMessage(ctx, cm)
	term.checkCommitted(ctx, blockHeight, view, blockHash)
}

func (term *LeanHelixTerm) HandleLeanHelixCommit(ctx context.Context, cm *CommitMessage) {
	term.logger.Debug("H=%s V=%s %s HandleLeanHelixCommit()", term.height, term.view, term.myMemberId.KeyForMap())
	header := cm.content.SignedHeader()
	sender := cm.content.Sender()

	if !term.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		term.logger.Debug("verification failed for Commit blockHeight=%v view=%v blockHash=%v", header.BlockHeight(), header.View(), header.BlockHash())
		return
	}
	term.storage.StoreCommit(cm)
	term.checkCommitted(ctx, header.BlockHeight(), header.View(), header.BlockHash())
}

func (term *LeanHelixTerm) checkCommitted(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	term.logger.Debug("H=%s V=%s %s checkCommitted() H=%s V=%s BlockHash %s ", term.height, term.view, term.myMemberId.KeyForMap(), blockHeight, view, blockHash)
	if term.committedBlock != nil {
		return
	}
	if !term.isPreprepared(blockHeight, view, blockHash) {
		return
	}
	commits := term.storage.GetCommitSendersIds(blockHeight, view, blockHash)
	if len(commits) < term.QuorumSize() {
		return
	}
	ppm, ok := term.storage.GetPreprepareMessage(blockHeight, view)
	if !ok {
		// log
		term.logger.Info("H=%s V=%s checkCommitted() missing PPM")
		return
	}
	term.logger.Info("H=%s V=%s %s checkCommitted() COMMITTED H=%s V=%s BlockHash %s ", term.height, term.view, term.myMemberId.KeyForMap(), blockHeight, view, blockHash)
	term.committedBlock = ppm.block
	term.onCommit(ctx, ppm.block, nil)
}

func (term *LeanHelixTerm) validateViewChangeVotes(targetBlockHeight primitives.BlockHeight, targetView primitives.View, confirmations []*protocol.ViewChangeMessageContent) bool {
	if len(confirmations) < term.QuorumSize() {
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

func (term *LeanHelixTerm) latestViewChangeVote(confirmations []*protocol.ViewChangeMessageContent) *protocol.ViewChangeMessageContent {
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

func (term *LeanHelixTerm) HandleLeanHelixNewView(ctx context.Context, nvm *NewViewMessage) {
	term.logger.Debug("H=%s V=%s %s HandleLeanHelixNewView()", term.height, term.view, term.myMemberId.KeyForMap())
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

	if !term.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", ignored because the signature verification failed` });
		term.logger.Debug("HandleLeanHelixNewView(): verify failed")
		return
	}

	futureLeaderId := term.calcLeaderMemberId(header.View())
	if !sender.MemberId().Equal(futureLeaderId) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", rejected because it match the new id (${view})` });
		term.logger.Debug("HandleLeanHelixNewView(): no match for future leader")
		return
	}

	if !term.validateViewChangeVotes(header.BlockHeight(), header.View(), viewChangeConfirmations) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", votes is invalid` });
		term.logger.Debug("HandleLeanHelixNewView(): validateViewChangeVotes failed")
		return
	}

	if term.view > header.View() {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", view is from the past` });
		term.logger.Debug("HandleLeanHelixNewView(): current view is higher than message view")
		return
	}

	if !ppMessageContent.SignedHeader().View().Equal(header.View()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", view doesn't match PP.view` });
		term.logger.Debug("HandleLeanHelixNewView(): NewView.view and NewView.Preprepare.view do not match")
		return
	}

	if !ppMessageContent.SignedHeader().BlockHeight().Equal(header.BlockHeight()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", blockHeight doesn't match PP.blockHeight` });
		term.logger.Debug("HandleLeanHelixNewView(): NewView.BlockHeight and NewView.Preprepare.BlockHeight do not match")
		return
	}

	latestVote := term.latestViewChangeVote(viewChangeConfirmations)
	if latestVote != nil {
		viewChangeMessageValid := term.isViewChangeValid(futureLeaderId, header.View(), latestVote)
		if !viewChangeMessageValid {
			//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", view change votes are invalid` });
			term.logger.Debug("HandleLeanHelixNewView(): NewView.ViewChangeConfirmation (with latest view) is invalid")
			return
		}

		// rewrite this mess
		latestVoteBlockHash := latestVote.SignedHeader().PreparedProof().PreprepareBlockRef().BlockHash()
		if latestVoteBlockHash != nil {
			isValidDigest := term.blockUtils.ValidateBlockCommitment(header.BlockHeight(), nvm.Block(), latestVoteBlockHash)
			if !isValidDigest {
				//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleLeanHelixNewView from "${senderId}", the given _Block (PP._Block) doesn't match the best _Block from the VCProof` });
				term.logger.Debug("HandleLeanHelixNewView(): NewView.ViewChangeConfirmation (with latest view) is invalid")
				return
			}
		}
	}

	ppm := &PreprepareMessage{
		content: ppMessageContent,
		block:   nvm.Block(),
	}

	if err := term.validatePreprepare(ctx, ppm); err == nil {
		term.newViewLocally = header.View()
		term.SetView(header.View())
		term.processPreprepare(ctx, ppm)
	}
}

func (term *LeanHelixTerm) QuorumSize() int {
	committeeMembersCount := len(term.committeeMembersMemberIds)
	f := int(math.Floor(float64(committeeMembersCount-1) / 3))
	return committeeMembersCount - f
}

func (term *LeanHelixTerm) IsLeader() bool {
	return term.myMemberId.Equal(term.leaderMemberId)
}

func (term *LeanHelixTerm) countPrepared(height primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) int {
	return len(term.storage.GetPrepareSendersIds(height, view, blockHash))
}

func (term *LeanHelixTerm) isPreprepared(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) bool {
	ppm, ok := term.storage.GetPreprepareMessage(blockHeight, view)
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

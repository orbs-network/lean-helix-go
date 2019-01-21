package termincommittee

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/blockextractor"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/services/proofsvalidator"
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"runtime"
	"sort"
)

// The algorithm cannot function with less committee members
// because it cannot calculate the f number (where committee members are 3f+1)
// The only reason to set this manually in config below this limit is for internal tests
const LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS = 4

func Str(memberId primitives.MemberId) string {
	if memberId == nil {
		return ""
	}
	return memberId.String()[:6]
}

type OnInCommitteeCommitCallback func(ctx context.Context, block interfaces.Block, commitMessages []*interfaces.CommitMessage)

type TermInCommittee struct {
	keyManager                     interfaces.KeyManager
	communication                  interfaces.Communication
	storage                        interfaces.Storage
	electionTrigger                interfaces.ElectionTrigger
	blockUtils                     interfaces.BlockUtils
	onCommit                       OnInCommitteeCommitCallback
	messageFactory                 *messagesfactory.MessageFactory
	myMemberId                     primitives.MemberId
	committeeMembersMemberIds      []primitives.MemberId
	otherCommitteeMembersMemberIds []primitives.MemberId
	height                         primitives.BlockHeight
	view                           primitives.View
	preparedLocally                bool
	committedBlock                 interfaces.Block
	leaderMemberId                 primitives.MemberId
	newViewLocally                 primitives.View
	logger                         interfaces.Logger
	prevBlock                      interfaces.Block
	QuorumSize                     int
}

func NewTermInCommittee(
	ctx context.Context,
	config *interfaces.Config,
	messageFactory *messagesfactory.MessageFactory,
	committeeMembers []primitives.MemberId,
	blockHeight primitives.BlockHeight,
	prevBlock interfaces.Block,
	onCommit OnInCommitteeCommitCallback) *TermInCommittee {

	keyManager := config.KeyManager
	blockUtils := config.BlockUtils
	membership := config.Membership
	myMemberId := membership.MyMemberId()
	comm := config.Communication

	panicOnLessThanMinimumCommitteeMembers(committeeMembers)

	otherCommitteeMembers := make([]primitives.MemberId, 0)
	for _, member := range committeeMembers {
		if !member.Equal(myMemberId) {
			otherCommitteeMembers = append(otherCommitteeMembers, member)
		}
	}
	if config.Logger == nil {
		config.Logger = logger.NewSilentLogger()
	}

	if config.Storage == nil {
		config.Storage = storage.NewInMemoryStorage()
	}

	config.Logger.Debug("H=%d V=0 ID=%s NewTermInCommittee: committeeMembersCount=%d", blockHeight, Str(myMemberId), len(committeeMembers))

	result := &TermInCommittee{
		height:                         blockHeight,
		onCommit:                       onCommit,
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
		QuorumSize:                     quorum.CalcQuorumSize(len(committeeMembers)),
	}

	result.startTerm(ctx)
	return result
}

func panicOnLessThanMinimumCommitteeMembers(committeeMembers []primitives.MemberId) {
	if len(committeeMembers) < LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS {
		panic(fmt.Sprintf("LH Received only %d committee members, but the hard minimum is %d", len(committeeMembers), LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS))
	}
}

func (tic *TermInCommittee) startTerm(ctx context.Context) {
	tic.initView(ctx, 0)
	if tic.isLeader() {
		tic.logger.Debug("H=%d V=%d ID=%s startTerm() I AM THE LEADER", tic.height, tic.view, Str(tic.myMemberId))
		block, blockHash := tic.blockUtils.RequestNewBlockProposal(ctx, tic.height, tic.prevBlock)
		ppm := tic.messageFactory.CreatePreprepareMessage(tic.height, tic.view, block, blockHash)

		tic.storage.StorePreprepare(ppm)
		tic.logger.Debug("H=%d V=%d ID=%s LHFLOW SEND PREPREPARE", tic.height, tic.view, Str(tic.myMemberId))
		tic.sendConsensusMessage(ctx, ppm)
	}
}

func (tic *TermInCommittee) GetView() primitives.View {
	return tic.view
}

func (tic *TermInCommittee) SetView(ctx context.Context, view primitives.View) {
	if tic.view != view {
		tic.initView(ctx, view)
	}
}

func (tic *TermInCommittee) initView(ctx context.Context, view primitives.View) {
	tic.preparedLocally = false
	tic.view = view
	tic.leaderMemberId = tic.calcLeaderMemberId(view)
	tic.electionTrigger.RegisterOnElection(ctx, tic.height, tic.view, tic.moveToNextLeader)
	tic.logger.Debug("H=%d V=%d ID=%s initView() set leader to %s, incremented view to %s, goroutines#=%d", tic.height, tic.view, Str(tic.myMemberId), Str(tic.leaderMemberId), tic.view, runtime.NumGoroutine())
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
	tic.SetView(ctx, tic.view+1)
	tic.logger.Debug("H=%d V=%d ID=%s LHFLOW moveToNextLeader() newLeader=%s", tic.height, tic.view, Str(tic.myMemberId), Str(tic.leaderMemberId))
	preparedMessages := preparedmessages.ExtractPreparedMessages(tic.height, tic.storage, tic.QuorumSize)
	vcm := tic.messageFactory.CreateViewChangeMessage(tic.height, tic.view, preparedMessages)
	if tic.isLeader() {
		tic.logger.Debug("H=%d V=%d ID=%s moveToNextLeader() I AM THE LEADER", tic.height, tic.view, Str(tic.myMemberId))
		tic.storage.StoreViewChange(vcm)
		tic.checkElected(ctx, tic.height, tic.view)
	} else {
		tic.sendConsensusMessage(ctx, vcm)
	}
}

func (tic *TermInCommittee) sendConsensusMessage(ctx context.Context, message interfaces.ConsensusMessage) {
	tic.logger.Debug("H=%d V=%d ID=%s sendConsensusMessage() msgType=%v", tic.height, tic.view, Str(tic.myMemberId), message.MessageType())
	rawMessage := interfaces.CreateConsensusRawMessage(message)
	tic.communication.SendConsensusMessage(ctx, tic.otherCommitteeMembersMemberIds, rawMessage)
}

func (tic *TermInCommittee) HandlePrePrepare(ctx context.Context, ppm *interfaces.PreprepareMessage) {
	tic.logger.Debug("H=%d V=%d ID=%s LHFLOW HandlePrePrepare() sender=%s", tic.height, tic.view, Str(tic.myMemberId), Str(ppm.SenderMemberId()))
	if err := tic.validatePreprepare(ctx, ppm); err != nil {
		tic.logger.Debug("H=%d V=%d ID=%s HandlePrePrepare() err=%v", tic.height, tic.view, Str(tic.myMemberId), err)
	} else {
		tic.processPreprepare(ctx, ppm)
	}
}

func (tic *TermInCommittee) processPreprepare(ctx context.Context, ppm *interfaces.PreprepareMessage) {
	header := ppm.Content().SignedHeader()
	if tic.view != header.View() {
		tic.logger.Debug("H=%d V=%d ID=%s processPreprepare() message from incorrect view %d", tic.height, tic.view, Str(tic.myMemberId), header.View())
		return
	}

	pm := tic.messageFactory.CreatePrepareMessage(header.BlockHeight(), header.View(), header.BlockHash())
	tic.storage.StorePreprepare(ppm)
	tic.storage.StorePrepare(pm)
	tic.logger.Debug("H=%d V=%d ID=%s LHFLOW SEND PREPARE", tic.height, tic.view, Str(tic.myMemberId))
	tic.sendConsensusMessage(ctx, pm)
	tic.checkPrepared(ctx, header.BlockHeight(), header.View(), header.BlockHash())
}

func (tic *TermInCommittee) validatePreprepare(ctx context.Context, ppm *interfaces.PreprepareMessage) error {
	blockHeight := ppm.BlockHeight()
	view := ppm.View()
	if tic.hasPreprepare(blockHeight, view) {
		return fmt.Errorf("already received Preprepare for H=%d V=%d", blockHeight, view)
	}

	header := ppm.Content().SignedHeader()
	sender := ppm.Content().Sender()
	if !tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		return fmt.Errorf("verification failed for sender %s signature on header", Str(sender.MemberId()))
	}

	leaderMemberId := tic.calcLeaderMemberId(view)
	senderMemberId := sender.MemberId()
	if !senderMemberId.Equal(leaderMemberId) {
		// Log
		return fmt.Errorf("sender %s is not leader", Str(senderMemberId))
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

func (tic *TermInCommittee) HandlePrepare(ctx context.Context, pm *interfaces.PrepareMessage) {
	tic.logger.Debug("H=%d V=%d ID=%s LHFLOW HandlePrepare() sender=%s", tic.height, tic.view, Str(tic.myMemberId), Str(pm.SenderMemberId()))
	header := pm.Content().SignedHeader()
	sender := pm.Content().Sender()

	if !tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		tic.logger.Debug("H=%d V=%d ID=%s verification failed for Prepare block-height=%v view=%d block-hash=%s", tic.height, tic.view, Str(tic.myMemberId), header.BlockHeight(), header.View(), header.BlockHash())
		return
	}
	if tic.view > header.View() {
		tic.logger.Debug("H=%d V=%d ID=%s prepare view %v is less than current term's view %v", tic.height, tic.view, Str(tic.myMemberId), header.View(), tic.view)
		return
	}
	if tic.leaderMemberId.Equal(sender.MemberId()) {
		tic.logger.Debug("H=%d V=%d ID=%s prepare received from leader (only preprepare can be received from leader)", tic.height, tic.view, Str(tic.myMemberId))
		return
	}
	tic.storage.StorePrepare(pm)
	if tic.view == header.View() {
		tic.checkPrepared(ctx, header.BlockHeight(), header.View(), header.BlockHash())
	}
}

func (tic *TermInCommittee) HandleViewChange(ctx context.Context, vcm *interfaces.ViewChangeMessage) {
	tic.logger.Debug("H=%d V=%d ID=%s LHFLOW HandleViewChange() sender=%s", tic.height, tic.view, Str(tic.myMemberId), Str(vcm.SenderMemberId()))
	if !tic.isViewChangeValid(tic.myMemberId, tic.view, vcm.Content()) {
		return
	}

	header := vcm.Content().SignedHeader()
	if vcm.Block() != nil && header.PreparedProof() != nil {
		isValidDigest := tic.blockUtils.ValidateBlockCommitment(vcm.BlockHeight(), vcm.Block(), header.PreparedProof().PreprepareBlockRef().BlockHash())
		if !isValidDigest {
			tic.logger.Debug("H=%d V=%d ID=%s different block hashes for block provided with message, and the block provided by the PPM in the PreparedProof of the message", tic.height, tic.view, Str(tic.myMemberId))
			return
		}
	}

	tic.storage.StoreViewChange(vcm)
	tic.checkElected(ctx, header.BlockHeight(), header.View())
}

func (tic *TermInCommittee) isViewChangeValid(targetLeaderMemberId primitives.MemberId, currentView primitives.View, vcm *protocol.ViewChangeMessageContent) bool {
	header := vcm.SignedHeader()
	sender := vcm.Sender()
	vcmView := header.View()
	preparedProof := header.PreparedProof()

	if !tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		tic.logger.Debug("H=%d V=%d ID=%s isViewChangeValid(): VerifyConsensusMessage() failed", tic.height, tic.view, Str(tic.myMemberId))
		return false
	}

	if currentView > vcmView {
		tic.logger.Debug("H=%d V=%d ID=%s isViewChangeValid(): message view %s is older than current term's view %s", tic.height, currentView, Str(tic.myMemberId), vcmView, currentView)
		return false
	}

	if !proofsvalidator.ValidatePreparedProof(tic.height, vcmView, preparedProof, tic.QuorumSize, tic.keyManager, tic.committeeMembersMemberIds, func(view primitives.View) primitives.MemberId { return tic.calcLeaderMemberId(view) }) {
		tic.logger.Debug("H=%d V=%d ID=%s isViewChangeValid(): failed ValidatePreparedProof()", tic.height, tic.view, Str(tic.myMemberId))
		return false
	}

	futureLeaderMemberId := tic.calcLeaderMemberId(vcmView)
	if !targetLeaderMemberId.Equal(futureLeaderMemberId) {
		return false
	}

	return true

}

func (tic *TermInCommittee) checkElected(ctx context.Context, height primitives.BlockHeight, view primitives.View) {
	if tic.newViewLocally < view {
		vcms, ok := tic.storage.GetViewChangeMessages(height, view)
		minimumNodes := tic.QuorumSize
		if ok && len(vcms) >= minimumNodes {
			tic.onElected(ctx, view, vcms[:minimumNodes])
		}
	}
}

func (tic *TermInCommittee) onElected(ctx context.Context, view primitives.View, viewChangeMessages []*interfaces.ViewChangeMessage) {
	tic.newViewLocally = view
	tic.SetView(ctx, view)
	block, blockHash := blockextractor.GetLatestBlockFromViewChangeMessages(viewChangeMessages)
	if block == nil {
		block, blockHash = tic.blockUtils.RequestNewBlockProposal(ctx, tic.height, tic.prevBlock)
	}
	ppmContentBuilder := tic.messageFactory.CreatePreprepareMessageContentBuilder(tic.height, view, block, blockHash)
	ppm := tic.messageFactory.CreatePreprepareMessageFromContentBuilder(ppmContentBuilder, block)
	confirmations := interfaces.ExtractConfirmationsFromViewChangeMessages(viewChangeMessages)
	nvm := tic.messageFactory.CreateNewViewMessage(tic.height, view, ppmContentBuilder, confirmations, block)
	tic.storage.StorePreprepare(ppm)
	tic.sendConsensusMessage(ctx, nvm)
}

func (tic *TermInCommittee) checkPrepared(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	if tic.preparedLocally == false {
		if tic.isPreprepared(blockHeight, view, blockHash) {
			countPrepared := tic.countPrepared(blockHeight, view, blockHash)
			isPrepared := countPrepared >= tic.QuorumSize-1
			tic.logger.Debug("H=%d V=%d ID=%s PREPARED expected=%d got=%d isPrepared=%t", tic.height, tic.view, Str(tic.myMemberId), tic.QuorumSize-1, countPrepared, isPrepared)
			if isPrepared {
				tic.onPrepared(ctx, blockHeight, view, blockHash)
			}
		}
	}
}

func (tic *TermInCommittee) onPrepared(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	tic.preparedLocally = true
	cm := tic.messageFactory.CreateCommitMessage(blockHeight, view, blockHash)
	tic.storage.StoreCommit(cm)
	tic.logger.Debug("H=%d V=%d ID=%s call checkCommitted()", tic.height, tic.view, Str(tic.myMemberId))
	tic.sendConsensusMessage(ctx, cm)
	tic.checkCommitted(ctx, blockHeight, view, blockHash)
}

func (tic *TermInCommittee) HandleCommit(ctx context.Context, cm *interfaces.CommitMessage) {
	tic.logger.Debug("H=%d V=%d ID=%s LHFLOW HandleCommit() sender=%s", tic.height, tic.view, Str(tic.myMemberId), Str(cm.SenderMemberId()))
	header := cm.Content().SignedHeader()
	sender := cm.Content().Sender()

	if !tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender) {
		tic.logger.Debug("H=%d V=%d ID=%s verification failed for Commit block-height=%d view=%d block-hash=%s", tic.height, tic.view, Str(tic.myMemberId), header.BlockHeight(), header.View(), header.BlockHash())
		return
	}
	tic.storage.StoreCommit(cm)
	tic.checkCommitted(ctx, header.BlockHeight(), header.View(), header.BlockHash())
}

func (tic *TermInCommittee) checkCommitted(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	tic.logger.Debug("H=%d V=%d ID=%s checkCommitted() H=%d V=%d block-hash=%s ", tic.height, tic.view, Str(tic.myMemberId), blockHeight, view, blockHash)
	if tic.committedBlock != nil {
		return
	}
	if !tic.isPreprepared(blockHeight, view, blockHash) {
		return
	}
	commits, ok := tic.storage.GetCommitMessages(blockHeight, view, blockHash)
	if !ok || len(commits) < tic.QuorumSize {
		return
	}
	ppm, ok := tic.storage.GetPreprepareMessage(blockHeight, view)
	if !ok {
		// log
		tic.logger.Info("H=%d V=%d checkCommitted() missing PPM", tic.height, tic.view)
		return
	}
	tic.logger.Info("H=%d V=%d ID=%s checkCommitted() COMMITTED calling onCommit() with block H=%d V=%d block-hash=%s num-commit-messages=%d", tic.height, tic.view, Str(tic.myMemberId), blockHeight, view, blockHash, len(commits))
	tic.committedBlock = ppm.Block()
	tic.onCommit(ctx, ppm.Block(), commits)
}

func (tic *TermInCommittee) validateViewChangeVotes(targetBlockHeight primitives.BlockHeight, targetView primitives.View, confirmations []*protocol.ViewChangeMessageContent) bool {
	if len(confirmations) < tic.QuorumSize {
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

func (tic *TermInCommittee) HandleNewView(ctx context.Context, nvm *interfaces.NewViewMessage) {
	tic.logger.Debug("H=%d V=%d ID=%s LHFLOW HandleNewView() sender=%s", tic.height, tic.view, Str(tic.myMemberId), nvm.SenderMemberId())
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
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", ignored because the signature verification failed` });
		tic.logger.Debug("H=%d V=%d ID=%s HandleNewView(): verify failed", tic.height, tic.view, Str(tic.myMemberId))
		return
	}

	futureLeaderId := tic.calcLeaderMemberId(header.View())
	if !sender.MemberId().Equal(futureLeaderId) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", rejected because it match the new id (${view})` });
		tic.logger.Debug("H=%d V=%d ID=%s HandleNewView(): no match for future leader", tic.height, tic.view, Str(tic.myMemberId))
		return
	}

	if !tic.validateViewChangeVotes(header.BlockHeight(), header.View(), viewChangeConfirmations) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", votes is invalid` });
		tic.logger.Debug("H=%d V=%d ID=%s HandleNewView(): validateViewChangeVotes failed", tic.height, tic.view, Str(tic.myMemberId))
		return
	}

	if tic.view > header.View() {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", view is from the past` });
		tic.logger.Debug("H=%d V=%d ID=%s HandleNewView(): current view is higher than message view", tic.height, tic.view, Str(tic.myMemberId))
		return
	}

	if !ppMessageContent.SignedHeader().View().Equal(header.View()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", view doesn't match PP.view` });
		tic.logger.Debug("H=%d V=%d ID=%s HandleNewView(): NewView.view and NewView.Preprepare.view do not match", tic.height, tic.view, Str(tic.myMemberId))
		return
	}

	if !ppMessageContent.SignedHeader().BlockHeight().Equal(header.BlockHeight()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", blockHeight doesn't match PP.Block()Height` });
		tic.logger.Debug("H=%d V=%d ID=%s HandleNewView(): NewView.BlockHeight and NewView.Preprepare.BlockHeight do not match", tic.height, tic.view, Str(tic.myMemberId))
		return
	}

	latestVote := tic.latestViewChangeVote(viewChangeConfirmations)
	if latestVote != nil {
		viewChangeMessageValid := tic.isViewChangeValid(futureLeaderId, header.View(), latestVote)
		if !viewChangeMessageValid {
			//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", view change votes are invalid` });
			tic.logger.Debug("H=%d V=%d ID=%s HandleNewView(): NewView.ViewChangeConfirmation (with latest view) is invalid", tic.height, tic.view, Str(tic.myMemberId))
			return
		}

		// rewrite this mess
		latestVoteBlockHash := latestVote.SignedHeader().PreparedProof().PreprepareBlockRef().BlockHash()
		if latestVoteBlockHash != nil {
			isValidDigest := tic.blockUtils.ValidateBlockCommitment(header.BlockHeight(), nvm.Block(), latestVoteBlockHash)
			if !isValidDigest {
				//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", the given _Block (PP._Block) doesn't match the best _Block from the VCProof` });
				tic.logger.Debug("H=%d V=%d ID=%s HandleNewView(): NewView.ViewChangeConfirmation (with latest view) is invalid", tic.height, tic.view, Str(tic.myMemberId))
				return
			}
		}
	}

	ppm := interfaces.NewPreprepareMessage(ppMessageContent, nvm.Block())

	if err := tic.validatePreprepare(ctx, ppm); err == nil {
		tic.newViewLocally = header.View()
		tic.SetView(ctx, header.View())
		tic.processPreprepare(ctx, ppm)
	}
}

func (tic *TermInCommittee) isLeader() bool {
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
	ppmBlock := ppm.Block()
	if ppmBlock == nil {
		return false
	}

	ppmBlockHash := ppm.Content().SignedHeader().BlockHash()
	return ppmBlockHash.Equal(blockHash)
}

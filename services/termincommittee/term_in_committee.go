// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package termincommittee

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/instrumentation/metrics"
	"github.com/orbs-network/lean-helix-go/services/blockextractor"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/messagesfactory"
	"github.com/orbs-network/lean-helix-go/services/preparedmessages"
	"github.com/orbs-network/lean-helix-go/services/proofsvalidator"
	"github.com/orbs-network/lean-helix-go/services/quorum"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/pkg/errors"
	"runtime"
	"sort"
	"strings"
)

// The algorithm cannot function with less committee members
// because it cannot calculate the f number (where committee members are 3f+1)
// The only reason to set this manually in config below this limit is for internal tests
const LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS = 4

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
	preparedLocally                *preparedLocallyProps
	committedBlock                 interfaces.Block
	newViewLocally                 primitives.View
	logger                         L.LHLogger
	prevBlock                      interfaces.Block
	QuorumSize                     int
}

func NewTermInCommittee(
	ctx context.Context,
	log L.LHLogger,
	config *interfaces.Config,
	messageFactory *messagesfactory.MessageFactory,
	committeeMembers []primitives.MemberId,
	blockHeight primitives.BlockHeight,
	prevBlock interfaces.Block,
	canBeFirstLeader bool,
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
	if config.Storage == nil {
		config.Storage = storage.NewInMemoryStorage()
	}

	log.Debug(L.LC(blockHeight, 0, myMemberId), "NewTermInCommittee: committeeMembersCount=%d members=%s", len(committeeMembers), ToCommitteeMembersStr(committeeMembers))

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
		logger:                         log,
		QuorumSize:                     quorum.CalcQuorumSize(len(committeeMembers)),
	}

	result.startTerm(ctx, canBeFirstLeader)
	return result
}

func Str(memberId primitives.MemberId) string {
	return L.MemberIdToStr(memberId)
}

type OnInCommitteeCommitCallback func(ctx context.Context, block interfaces.Block, commitMessages []*interfaces.CommitMessage)

type preparedLocallyProps struct {
	isPreparedLocally bool
	latestView        primitives.View
}

func (tic *TermInCommittee) getPreparedLocally() (v primitives.View, ok bool) {
	if tic.preparedLocally == nil || !tic.preparedLocally.isPreparedLocally {
		return 0, false
	}
	return tic.preparedLocally.latestView, true
}

func (tic *TermInCommittee) setNotPreparedLocally() {
	tic.preparedLocally = nil
}

func (tic *TermInCommittee) setPreparedLocally(v primitives.View) {
	tic.preparedLocally = &preparedLocallyProps{
		isPreparedLocally: true,
		latestView:        v,
	}
}

func ToCommitteeMembersStr(members []primitives.MemberId) string {

	var strs []string
	for _, member := range members {
		strs = append(strs, Str(member))
	}
	return strings.Join(strs, ", ")
}

func panicOnLessThanMinimumCommitteeMembers(committeeMembers []primitives.MemberId) {
	if len(committeeMembers) < LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS {
		panic(fmt.Sprintf("LH Received only %d committee members, but the hard minimum is %d", len(committeeMembers), LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS))
	}
}

func (tic *TermInCommittee) startTerm(ctx context.Context, canBeFirstLeader bool) {
	tic.setNotPreparedLocally()
	tic.initView(ctx, 0)
	if tic.height > 1 && !canBeFirstLeader {
		tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW startTerm() I CANNOT BE LEADER OF FIRST VIEW, skipping view")
		return
	}
	if err := tic.isLeader(tic.myMemberId, tic.view); err != nil {
		tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW startTerm() I AM THE LEADER OF FIRST VIEW, requesting new block")
		block, blockHash := tic.blockUtils.RequestNewBlockProposal(ctx, tic.height, tic.prevBlock)
		ppm := tic.messageFactory.CreatePreprepareMessage(tic.height, tic.view, block, blockHash)

		tic.storage.StorePreprepare(ppm)
		tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG SEND PREPREPARE")
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
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW PreparedLocally set to false")
	tic.view = view
	tic.electionTrigger.RegisterOnElection(ctx, tic.height, tic.view, tic.moveToNextLeader)
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW initView() set leader to %s, incremented view to %d, election-timeout=%s, members=%s, goroutines#=%d", Str(tic.calcLeaderMemberId(view)), tic.view, tic.electionTrigger.CalcTimeout(view), ToCommitteeMembersStr(tic.committeeMembersMemberIds), runtime.NumGoroutine())
}

func (tic *TermInCommittee) Dispose() {
	tic.electionTrigger.Stop()
	tic.storage.ClearBlockHeightLogs(tic.height)
}

func (tic *TermInCommittee) calcLeaderMemberId(view primitives.View) primitives.MemberId {
	return calcLeaderOfViewAndCommittee(view, tic.committeeMembersMemberIds)
}

func calcLeaderOfViewAndCommittee(view primitives.View, committeeMembersMemberIds []primitives.MemberId) primitives.MemberId {
	index := int(view) % len(committeeMembersMemberIds)
	return committeeMembersMemberIds[index]
}

func (tic *TermInCommittee) moveToNextLeader(ctx context.Context, height primitives.BlockHeight, view primitives.View, onElectionCB func(m metrics.ElectionMetrics)) {
	if view != tic.view || height != tic.height {
		return
	}
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW moveToNextLeader() calling SetView(), incrementing view to V=%d", tic.view+1)
	tic.SetView(ctx, tic.view+1)
	newLeader := tic.calcLeaderMemberId(tic.view)
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW moveToNextLeader() newLeader=%s", Str(newLeader))
	var preparedMessages *preparedmessages.PreparedMessages
	if tic.preparedLocally != nil && tic.preparedLocally.isPreparedLocally {
		preparedMessages = preparedmessages.ExtractPreparedMessages(tic.height, tic.preparedLocally.latestView, tic.storage, tic.QuorumSize)
	}
	vcm := tic.messageFactory.CreateViewChangeMessage(tic.height, tic.view, preparedMessages)

	if err := tic.isLeader(tic.myMemberId, tic.view); err == nil {
		tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW moveToNextLeader() I will be leader if I get enough VIEW_CHANGE votes. My leadership of V=%d will time out in %s", tic.view, tic.electionTrigger.CalcTimeout(tic.view))
		tic.storage.StoreViewChange(vcm)
		tic.checkElected(ctx, tic.height, tic.view)
	} else {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG SEND VIEW_CHANGE (I'm not leader) moveToNextLeader() (%s)", err)
		tic.sendConsensusMessageToSpecificMember(ctx, newLeader, vcm)
	}
	if onElectionCB != nil {
		onElectionCB(metrics.NewElectionMetrics(newLeader, tic.view))
	}
}

// TODO Consider returning error with who is the expected leader (tic.calcLeaderMemberId(v)), to help caller debug this
func (tic *TermInCommittee) isLeader(memberId primitives.MemberId, v primitives.View) error {
	return isLeaderOfViewForThisCommittee(memberId, v, tic.committeeMembersMemberIds)
}

func isLeaderOfViewForThisCommittee(leaderCandidate primitives.MemberId, v primitives.View, committeeMembersMemberIds []primitives.MemberId) error {

	calculatedLeader := calcLeaderOfViewAndCommittee(v, committeeMembersMemberIds)
	if !leaderCandidate.Equal(calculatedLeader) {
		return errors.Errorf("candidate leader is %s but calculated leader for V=%s is %s", Str(leaderCandidate), v, Str(calculatedLeader))
	}
	return nil
}

func (tic *TermInCommittee) checkElected(ctx context.Context, height primitives.BlockHeight, view primitives.View) {
	if tic.newViewLocally >= view {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "checkElected() already newViewLocally=%d is greater or equal to received view=%d, skipping", tic.newViewLocally, view)
		return
	}
	vcms, ok := tic.storage.GetViewChangeMessages(height, view)
	minimumNodes := tic.QuorumSize
	if !ok {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "checkElected() could not get stored VIEW_CHANGE messages, skipping")
		return
	}

	if len(vcms) < minimumNodes {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "checkElected() stored %d of %d VIEW_CHANGE messages (last-sender=%s)", len(vcms), minimumNodes, Str(vcms[len(vcms)-1].SenderMemberId()))
		return
	}
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "checkElected() stored %d of %d VIEW_CHANGE messages (last-sender=%s)", len(vcms), minimumNodes, Str(vcms[len(vcms)-1].SenderMemberId()))
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "checkElected() has enough VIEW_CHANGE messages, proceeding to onElected() with V=%d", view)
	tic.onElected(ctx, view, vcms[:minimumNodes])
}

func (tic *TermInCommittee) onElected(ctx context.Context, view primitives.View, viewChangeMessages []*interfaces.ViewChangeMessage) {
	tic.newViewLocally = view
	tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW onElected() I AM THE LEADER BY VIEW CHANGE for V=%d, now calling SetView()", view)
	tic.SetView(ctx, view)
	block, blockHash := blockextractor.GetLatestBlockFromViewChangeMessages(viewChangeMessages)
	if block == nil {
		tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW onElected() MISSING BLOCK IN VIEW_CHANGE, calling RequestNewBlockProposal()")
		block, blockHash = tic.blockUtils.RequestNewBlockProposal(ctx, tic.height, tic.prevBlock)
		tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW onElected() returned from RequestNewBlockProposal(), sending the new block as part of NEW_VIEW")
	} else {
		tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW onElected() found block with H=%d in VIEW_CHANGE messages, so sending it as part of NEW_VIEW", block.Height())
	}
	ppmContentBuilder := tic.messageFactory.CreatePreprepareMessageContentBuilder(tic.height, view, block, blockHash)
	ppm := tic.messageFactory.CreatePreprepareMessageFromContentBuilder(ppmContentBuilder, block)
	confirmations := interfaces.ExtractConfirmationsFromViewChangeMessages(viewChangeMessages)
	nvm := tic.messageFactory.CreateNewViewMessage(tic.height, view, ppmContentBuilder, confirmations, block)
	tic.storage.StorePreprepare(ppm)
	tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG SEND NEW_VIEW")
	tic.sendConsensusMessage(ctx, nvm)
}

func (tic *TermInCommittee) sendConsensusMessage(ctx context.Context, message interfaces.ConsensusMessage) {
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "sendConsensusMessage() msgType=%v", message.MessageType())
	rawMessage := interfaces.CreateConsensusRawMessage(message)
	tic.communication.SendConsensusMessage(ctx, tic.otherCommitteeMembersMemberIds, rawMessage)
}

func (tic *TermInCommittee) sendConsensusMessageToSpecificMember(ctx context.Context, targetMemberId primitives.MemberId, message interfaces.ConsensusMessage) {
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "sendConsensusMessageToSpecificMember() target=%s, msgType=%v", Str(targetMemberId), message.MessageType())
	rawMessage := interfaces.CreateConsensusRawMessage(message)
	tic.communication.SendConsensusMessage(ctx, []primitives.MemberId{targetMemberId}, rawMessage)
}

func (tic *TermInCommittee) HandlePrePrepare(ctx context.Context, ppm *interfaces.PreprepareMessage) {
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPREPARE (H=%d V=%d sender=%s)", ppm.BlockHeight(), ppm.View(), Str(ppm.SenderMemberId()))

	if err := tic.validatePreprepare(ctx, ppm); err != nil {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPREPARE IGNORE: validatePreprepare() failed: %s", err)
		return
	}

	err := tic.blockUtils.ValidateBlockProposal(ctx, ppm.BlockHeight(), ppm.Block(), ppm.Content().SignedHeader().BlockHash(), tic.prevBlock)
	if err != nil {
		tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPREPARE IGNORE: blockUtils.ValidateBlockProposal() failed: %s", err)
	}

	tic.processPreprepare(ctx, ppm)

}

func (tic *TermInCommittee) validatePreprepare(ctx context.Context, ppm *interfaces.PreprepareMessage) error {
	blockHeight := ppm.BlockHeight()
	if tic.hasPreprepare(blockHeight, ppm.View()) {
		errMsg := fmt.Sprintf("already stored Preprepare for H=%d V=%d", blockHeight, ppm.View())
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPREPARE IGNORE: hasPreprepare() failed: %s", errMsg)
		return errors.New(errMsg)
	}

	header := ppm.Content().SignedHeader()
	sender := ppm.Content().Sender()
	if err := tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender); err != nil {
		return errors.Wrapf(err, "verification failed for sender %s signature on header", Str(sender.MemberId()))
	}

	if err := tic.isLeader(sender.MemberId(), ppm.View()); err != nil {
		return fmt.Errorf("PREPREPARE sender %s is not leader: %s", Str(sender.MemberId()), err)
	}

	return nil
}

func (tic *TermInCommittee) hasPreprepare(blockHeight primitives.BlockHeight, view primitives.View) bool {
	_, ok := tic.storage.GetPreprepareMessage(blockHeight, view)
	return ok
}

func (tic *TermInCommittee) processPreprepare(ctx context.Context, ppm *interfaces.PreprepareMessage) {
	// TODO per spec move this to validatePreprepare()
	header := ppm.Content().SignedHeader()
	if tic.view != header.View() {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "processPreprepare() message from incorrect view %d", header.View())
		return
	}

	pm := tic.messageFactory.CreatePrepareMessage(header.BlockHeight(), header.View(), header.BlockHash())
	tic.storage.StorePreprepare(ppm)
	tic.storage.StorePrepare(pm)
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG SEND PREPARE")
	tic.sendConsensusMessage(ctx, pm)

	if err := tic.checkPreparedLocally(ctx, header.BlockHeight(), header.View(), header.BlockHash()); err != nil {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "checkPreparedLocally: err=%v", err)
	}
}

func (tic *TermInCommittee) HandlePrepare(ctx context.Context, pm *interfaces.PrepareMessage) {
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPARE (H=%d V=%d sender=%s)", pm.BlockHeight(), pm.View(), Str(pm.SenderMemberId()))
	header := pm.Content().SignedHeader()
	sender := pm.Content().Sender()

	if err := tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender); err != nil {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPARE IGNORE - verification failed for Prepare block-height=%v view=%d block-hash=%s err=%v", header.BlockHeight(), header.View(), header.BlockHash(), err)
		return
	}
	if header.View() < tic.view {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPARE IGNORE - prepare view %v is less than current term's view %v", header.View(), tic.view)
		return
	}
	if err := tic.isLeader(sender.MemberId(), header.View()); err == nil {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPARE IGNORE - prepare received from leader (only preprepare can be received from leader)")
		return
	}
	tic.storage.StorePrepare(pm)
	if header.View() > tic.view {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPARE STORE in future view %d", header.View())
	}
	if err := tic.checkPreparedLocally(ctx, header.BlockHeight(), header.View(), header.BlockHash()); err != nil {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "checkPreparedLocally: err=%v", err)
	}
}

func (tic *TermInCommittee) checkPreparedLocally(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) error {
	v, ok := tic.getPreparedLocally()
	if ok && v == view {
		return errors.Errorf("already in PHASE PREPARED for V=%d", view)
	}

	if err := tic.isPreprepared(blockHeight, view, blockHash); err != nil {
		return errors.Wrap(err, "isPreprepared failed")
	}

	countPrepared := tic.countPrepared(blockHeight, view, blockHash)
	isPrepared := countPrepared >= tic.QuorumSize-1
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW Check if in PHASE PREPARED: stored=%d out of expected=%d isPrepared=%t", countPrepared, tic.QuorumSize-1, isPrepared)
	if isPrepared {
		tic.onPreparedLocally(ctx, blockHeight, view, blockHash)
	}
	return nil
}

func (tic *TermInCommittee) isPreprepared(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) error {
	ppm, ok := tic.storage.GetPreprepareMessage(blockHeight, view)
	if !ok {
		return errors.New("PREPREPARE is not stored")
	}
	ppmBlock := ppm.Block()
	if ppmBlock == nil {
		return errors.New("Stored PREPREPARE does not contain a block")
	}

	ppmBlockHash := ppm.Content().SignedHeader().BlockHash()
	if !ppmBlockHash.Equal(blockHash) {
		return errors.New("Stored PREPREPARE blockHash is different from provided")
	}
	return nil
}

func (tic *TermInCommittee) countPrepared(height primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) int {
	return len(tic.storage.GetPrepareSendersIds(height, view, blockHash))
}

func (tic *TermInCommittee) onPreparedLocally(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	tic.setPreparedLocally(view)
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW PHASE PREPARED, PreparedLocally set to V=%d", view)
	cm := tic.messageFactory.CreateCommitMessage(blockHeight, view, blockHash)
	tic.storage.StoreCommit(cm)
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG SEND COMMIT")
	tic.sendConsensusMessage(ctx, cm)
	tic.checkCommitted(ctx, blockHeight, view, blockHash)
}

func (tic *TermInCommittee) HandleCommit(ctx context.Context, cm *interfaces.CommitMessage) {
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED COMMIT (H=%d V=%d sender=%s)", cm.BlockHeight(), cm.View(), Str(cm.SenderMemberId()))
	header := cm.Content().SignedHeader()
	sender := cm.Content().Sender()

	if err := tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender); err != nil {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED COMMIT IGNORE - verification failed for Commit block-height=%d view=%d block-hash=%s err=%v", header.BlockHeight(), header.View(), header.BlockHash(), err)
		return
	}
	tic.storage.StoreCommit(cm)
	tic.checkCommitted(ctx, header.BlockHeight(), header.View(), header.BlockHash())
}

func (tic *TermInCommittee) checkCommitted(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	if tic.committedBlock != nil {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED COMMIT IGNORE - already committed")
		return
	}
	if err := tic.isPreprepared(blockHeight, view, blockHash); err != nil {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED COMMIT IGNORE - is not preprepared, err=%v", err)
		return
	}
	commits, ok := tic.storage.GetCommitMessages(blockHeight, view, blockHash)
	if !ok || len(commits) < tic.QuorumSize {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED COMMIT STORE - received %d of %d required quorum commits", len(commits), tic.QuorumSize)
		return
	}
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED COMMIT STORE - received %d of %d required quorum commits", len(commits), tic.QuorumSize)
	ppm, ok := tic.storage.GetPreprepareMessage(blockHeight, view)
	if !ok {
		// log
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED COMMIT IGNORE - missing PPM in Commit message")
		return
	}
	var iSentCommitMessage bool
	for _, msg := range commits {
		if msg.SenderMemberId().Equal(tic.myMemberId) {
			iSentCommitMessage = true
			break
		}
	}
	if !iSentCommitMessage {
		cm := tic.messageFactory.CreateCommitMessage(blockHeight, view, blockHash)
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG SEND COMMIT because I did not send it during onPreparedLocally")
		tic.sendConsensusMessage(ctx, cm)
	}
	tic.committedBlock = ppm.Block()
	tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW PHASE COMMITTED CommittedBlock set to H=%d, calling onCommit() with H=%d V=%d block-hash=%s num-commit-messages=%d", ppm.Block().Height(), blockHeight, view, blockHash, len(commits))
	tic.onCommit(ctx, ppm.Block(), commits)
}

func (tic *TermInCommittee) HandleViewChange(ctx context.Context, vcm *interfaces.ViewChangeMessage) {
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED VIEW_CHANGE with H=%d V=%d sender=%s", vcm.BlockHeight(), vcm.View(), Str(vcm.SenderMemberId()))

	// isViewChangeIgnored() - contains calculatedLeaderFromViewChange and old view
	// Put the above code in isViewChangeIgnored() and call it in line 638 instead of the code there.
	// Do not accept VIEW_CHANGE from the past - this is not an error, just ignore it

	if err := tic.isViewChangeAccepted(tic.myMemberId, tic.view, vcm.Content()); err != nil {
		tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED VIEW_CHANGE IGNORE - %s", err)
		return
	}

	if err := tic.isViewChangeValid(tic.myMemberId, tic.view, vcm.Content()); err != nil {
		tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED VIEW_CHANGE IGNORE - invalid VIEW_CHANGE: %s", err)
		return
	}

	header := vcm.Content().SignedHeader()
	if vcm.Block() != nil && header.PreparedProof() != nil {
		isValidDigest := tic.blockUtils.ValidateBlockCommitment(vcm.BlockHeight(), vcm.Block(), header.PreparedProof().PreprepareBlockRef().BlockHash())
		if !isValidDigest {
			tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED VIEW_CHANGE IGNORE - different block hashes for block provided with message, and the block provided by the PPM in the PreparedProof of the message")
			return
		}
	}

	tic.storage.StoreViewChange(vcm)
	tic.checkElected(ctx, header.BlockHeight(), header.View())
}

func (tic *TermInCommittee) isViewChangeAccepted(expectedLeaderForView primitives.MemberId, view primitives.View, vcmContent *protocol.ViewChangeMessageContent) error {
	vcmView := vcmContent.SignedHeader().View()
	calculatedLeaderForView := tic.calcLeaderMemberId(vcmView)
	if !expectedLeaderForView.Equal(calculatedLeaderForView) {
		return errors.Errorf("I am not the calculated leader %s who should collect these messages - I am %s", Str(calculatedLeaderForView), Str(expectedLeaderForView))
	}

	if view > vcmView {
		return errors.Errorf("message view %s is older than current term's view %s", vcmView, view)
	}
	return nil
}

func (tic *TermInCommittee) isViewChangeValid(expectedLeaderFromNewView primitives.MemberId, currentView primitives.View, vcm *protocol.ViewChangeMessageContent) error {
	header := vcm.SignedHeader()
	sender := vcm.Sender()
	vcmView := header.View()
	preparedProof := header.PreparedProof()

	if err := tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender); err != nil {
		return errors.Wrapf(err, "keyManager.VerifyConsensusMessage failed")
	}

	if !proofsvalidator.ValidatePreparedProof(tic.height, vcmView, preparedProof, tic.QuorumSize, tic.keyManager, tic.committeeMembersMemberIds, func(view primitives.View) primitives.MemberId { return tic.calcLeaderMemberId(view) }) {
		return fmt.Errorf("failed ValidatePreparedProof()")
	}
	return nil
}

func (tic *TermInCommittee) validateViewChangeVotes(targetBlockHeight primitives.BlockHeight, targetView primitives.View, confirmations []*protocol.ViewChangeMessageContent) error {
	if len(confirmations) < tic.QuorumSize {
		return fmt.Errorf("there are %d confirmations but %d are needed", len(confirmations), tic.QuorumSize)
	}

	set := make(map[string]bool)

	// VerifyConsensusMessage that all _Block heights and views match, and all public keys are unique
	for _, confirmation := range confirmations {
		senderMemberIdStr := string(confirmation.Sender().MemberId())
		confirmationBlockHeight := confirmation.SignedHeader().BlockHeight()
		if confirmationBlockHeight != targetBlockHeight {
			return fmt.Errorf("confirmation of memberId %s has block height %d which is different than targetBlockHeight %d ",
				senderMemberIdStr, confirmationBlockHeight, targetBlockHeight)
		}
		confirmationView := confirmation.SignedHeader().View()
		if confirmationView != targetView {
			return fmt.Errorf("confirmation of memberId %s has view %d which is different than targetView %d ",
				senderMemberIdStr, confirmationView, targetView)
		}
		if set[senderMemberIdStr] {
			return fmt.Errorf("memberId %s appears in more than one confirmation", senderMemberIdStr)
		}
		set[senderMemberIdStr] = true
	}

	return nil

}

func (tic *TermInCommittee) HandleNewView(ctx context.Context, nvm *interfaces.NewViewMessage) {
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW (H=%d V=%d sender=%s)", nvm.BlockHeight(), nvm.View(), Str(nvm.SenderMemberId()))
	nvmHeader := nvm.Content().SignedHeader()
	nvmSender := nvm.Content().Sender()
	ppMessageContent := nvm.Content().Message()
	viewChangeConfirmationsIter := nvmHeader.ViewChangeConfirmationsIterator()
	viewChangeConfirmations := make([]*protocol.ViewChangeMessageContent, 0, 1)
	for {
		if !viewChangeConfirmationsIter.HasNext() {
			break
		}
		viewChangeConfirmations = append(viewChangeConfirmations, viewChangeConfirmationsIter.NextViewChangeConfirmations())
	}

	if err := tic.keyManager.VerifyConsensusMessage(nvmHeader.BlockHeight(), nvmHeader.Raw(), nvmSender); err != nil {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], handleNewViewMessage from "${senderId}", ignored because the signature verification failed` });
		tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - keyManager.VerifyConsensusMessage() failed: %s", err)
		return
	}

	calculatedLeaderFromNewView := tic.calcLeaderMemberId(nvmHeader.View())
	if err := tic.isLeader(nvmSender.MemberId(), nvmHeader.View()); err != nil {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], handleNewViewMessage from "${senderId}", rejected because it match the new id (${view})` });
		tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - sender %s no match for future leader: %s", nvmSender.MemberId(), err)
		return
	}

	if err := tic.validateViewChangeVotes(nvmHeader.BlockHeight(), nvmHeader.View(), viewChangeConfirmations); err != nil {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], handleNewViewMessage from "${senderId}", votes is invalid` });
		tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - validateViewChangeVotes failed: %s", err)
		return
	}

	if tic.view > nvmHeader.View() {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], handleNewViewMessage from "${senderId}", view is from the past` });
		tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - current view %d is higher than message view %d", tic.view, nvmHeader.View())
		return
	}

	ppmView := ppMessageContent.SignedHeader().View()
	if !ppmView.Equal(nvmHeader.View()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], handleNewViewMessage from "${senderId}", view doesn't match PP.view` });
		tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - NewView.view %d and NewView.Preprepare.view %d do not match",
			nvmHeader.View(), ppmView)
		return
	}

	if !ppMessageContent.SignedHeader().BlockHeight().Equal(nvmHeader.BlockHeight()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], handleNewViewMessage from "${senderId}", blockHeight doesn't match PP.Block()Height` });
		tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - NewView.BlockHeight and NewView.Preprepare.BlockHeight do not match")
		return
	}

	latestVote := tic.latestViewChangeVote(viewChangeConfirmations)
	if latestVote != nil {

		calculatedLeaderFromViewChange := tic.calcLeaderMemberId(latestVote.SignedHeader().View())
		if !calculatedLeaderFromNewView.Equal(calculatedLeaderFromViewChange) {
			tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - calculatedLeaderFromNewView=%s is not the calculated leader %s who should collect these messages", Str(calculatedLeaderFromNewView), Str(calculatedLeaderFromViewChange))
			return
		}

		if err := tic.isViewChangeValid(calculatedLeaderFromNewView, nvmHeader.View(), latestVote); err != nil {
			tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - NewView.ViewChangeConfirmation (with latest view) is invalid: %s", err)
			return
		}

		// rewrite this mess
		latestVoteBlockHash := latestVote.SignedHeader().PreparedProof().PreprepareBlockRef().BlockHash()
		if latestVoteBlockHash != nil {
			isValidDigest := tic.blockUtils.ValidateBlockCommitment(nvmHeader.BlockHeight(), nvm.Block(), latestVoteBlockHash)
			if !isValidDigest {
				//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], handleNewViewMessage from "${senderId}", the given _Block (PP._Block) doesn't match the best _Block from the VCProof` });
				tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - NewView.ViewChangeConfirmation (with latest view) is invalid")
				return
			}
		}
	}

	ppm := interfaces.NewPreprepareMessage(ppMessageContent, nvm.Block())

	// leader proposed a new block in this view, checking its proposal
	if latestVote == nil {
		err := tic.blockUtils.ValidateBlockProposal(ctx, ppm.BlockHeight(), ppm.Block(), ppm.Content().SignedHeader().BlockHash(), tic.prevBlock)
		if err != nil {
			tic.logger.Error(L.LC(tic.height, tic.view, tic.myMemberId), "Proposed block failed ValidateBlockProposal: %s", err)
			return
		}
	}

	if err := tic.validatePreprepare(ctx, ppm); err == nil {
		tic.newViewLocally = nvmHeader.View()
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW NewViewLocally set to V=%d (handleNewViewMessage), calling SetView()", tic.newViewLocally)
		tic.SetView(ctx, nvmHeader.View())
		tic.processPreprepare(ctx, ppm)
	} else {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW LHMSG RECEIVED NEW_VIEW FAILED validation of PPM: %s", err.Error())
	}
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

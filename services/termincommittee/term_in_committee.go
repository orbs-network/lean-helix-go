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
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/pkg/errors"
	"math"
	"runtime"
	"sort"
	"strings"
)

// The algorithm cannot function with less committee members
// because it cannot calculate the f number (where committee members are 3f+1)
// The only reason to set this manually in config below this limit is for internal tests
const LeanHelixHardMinimumCommitteeMembers = 4
const MaxView = math.MaxUint64

type TermInCommittee struct {
	keyManager                      interfaces.KeyManager
	communication                   interfaces.Communication
	storage                         interfaces.Storage
	electionTrigger                 interfaces.ElectionScheduler
	blockUtils                      interfaces.BlockUtils
	onCommit                        OnInCommitteeCommitCallback
	messageFactory                  *messagesfactory.MessageFactory
	myMemberId                      primitives.MemberId
	committeeMembersMemberIds       []primitives.MemberId
	otherCommitteeMembersMemberIds  []primitives.MemberId
	preparedLocally                 *preparedLocallyProps
	latestViewThatProcessedVCMOrNVM primitives.View
	committedBlock                  interfaces.Block
	logger                          L.LHLogger
	prevBlock                       interfaces.Block
	QuorumSize                      int
	State                           *state.State
}

func NewTermInCommittee(log L.LHLogger, config *interfaces.Config, state *state.State, messageFactory *messagesfactory.MessageFactory, electionTrigger interfaces.ElectionScheduler, committeeMembers []primitives.MemberId, prevBlock interfaces.Block, canBeFirstLeader bool, onCommit OnInCommitteeCommitCallback) *TermInCommittee {

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

	log.Debug("NewTermInCommittee: committeeMembersCount=%d members=%s", len(committeeMembers), ToCommitteeMembersStr(committeeMembers))

	result := &TermInCommittee{
		State:                          state,
		onCommit:                       onCommit,
		prevBlock:                      prevBlock,
		keyManager:                     keyManager,
		communication:                  comm,
		storage:                        config.Storage,
		electionTrigger:                electionTrigger,
		blockUtils:                     blockUtils,
		committeeMembersMemberIds:      committeeMembers,
		otherCommitteeMembersMemberIds: otherCommitteeMembers,
		messageFactory:                 messageFactory,
		myMemberId:                     myMemberId,
		logger:                         log,
		QuorumSize:                     quorum.CalcQuorumSize(len(committeeMembers)),
	}

	result.startTerm(canBeFirstLeader)
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
	if len(committeeMembers) < LeanHelixHardMinimumCommitteeMembers {
		panic(fmt.Sprintf("LH Received only %d committee members, but the hard minimum is %d", len(committeeMembers), LeanHelixHardMinimumCommitteeMembers))
	}
}

func (tic *TermInCommittee) startTerm(canBeFirstLeader bool) {
	tic.setNotPreparedLocally()

	currentHV, err := tic.initView(0)
	if err != nil {
		tic.logger.Info("LHFLOW startTerm() tried to SetView(0) while in state %s. failed: %s", currentHV, err)
		return
	}

	if currentHV.Height() > 1 && !canBeFirstLeader {
		tic.logger.Info("LHFLOW startTerm() I CANNOT BE LEADER OF FIRST VIEW, skipping view")
		return
	}

	if err := tic.isLeader(tic.myMemberId, currentHV.View()); err != nil {
		return // not leader, do nothing
	}

	tic.logger.Debug("LHFLOW startTerm() I AM THE LEADER OF FIRST VIEW, requesting new block")

	ctx, err := tic.State.Contexts.For(currentHV)
	if err != nil {
		tic.logger.Info("LHFLOW onElectedByViewChange() not requesting new block - %e", err)
		return
	}

	block, blockHash := tic.blockUtils.RequestNewBlockProposal(ctx, currentHV.Height(), tic.myMemberId, tic.prevBlock)

	// Sometimes PPM will still be sent although context was canceled,
	// because cancellation is not fast enough.
	// Context cancellation is only a performance optimization,
	// so whether PPM is sent out or not, does not affect correctness
	if ctx.Err() != nil {
		tic.logger.Info("LHFLOW startTerm() RequestNewBlockProposal() context canceled, not sending PREPREPARE - %s", ctx.Err())
		return
	}

	ppm := tic.messageFactory.CreatePreprepareMessage(currentHV.Height(), currentHV.View(), block, blockHash)

	tic.storage.StorePreprepare(ppm)
	tic.logger.Debug("LHMSG SEND PREPREPARE (msg: H=%d V=%d sender=%s)",
		ppm.BlockHeight(), ppm.View(), Str(ppm.SenderMemberId()))
	if err := tic.sendConsensusMessage(ppm); err != nil {
		tic.logger.Info("LHMSG SEND PREPREPARE FAILED - %s", err)
	}
}

// update view and reset election trigger
func (tic *TermInCommittee) initView(newView primitives.View) (*state.HeightView, error) {

	// Updates the state
	current, err := tic.State.SetView(newView)
	if err != nil {
		tic.logger.Info("LHFLOW initView() tried to SetView(%d) while in state %s. failed: %s", newView, current, err)
		return nil, err
	}

	tic.electionTrigger.RegisterOnElection(current.Height(), current.View(), tic.moveToNextLeaderByElection)
	tic.logger.Debug("LHFLOW initView() set leader to %s, incremented view to %d, election-timeout=%s, members=%s, goroutines#=%d",
		Str(tic.calcLeaderMemberId(current.View())), current.View(), tic.electionTrigger.CalcTimeout(current.View()),
		ToCommitteeMembersStr(tic.committeeMembersMemberIds), runtime.NumGoroutine())

	return current, nil
}

func (tic *TermInCommittee) Dispose() {
	tic.electionTrigger.Stop()
	height := tic.State.Height()
	tic.storage.ClearBlockHeightLogs(height)
	tic.logger.Debug("LHFLOW Dispose() for H=%d", height)
}

func (tic *TermInCommittee) calcLeaderMemberId(view primitives.View) primitives.MemberId {
	return calcLeaderOfViewAndCommittee(view, tic.committeeMembersMemberIds)
}

func calcLeaderOfViewAndCommittee(view primitives.View, committeeMembersMemberIds []primitives.MemberId) primitives.MemberId {
	index := int(view) % len(committeeMembersMemberIds)
	return committeeMembersMemberIds[index]
}

func (tic *TermInCommittee) moveToNextLeaderByElection(height primitives.BlockHeight, view primitives.View, updateMetrics interfaces.OnElectionCallback) {

	currentHV := tic.State.HeightView()
	if height != currentHV.Height() || view != currentHV.View() {
		return
	}
	tic.logger.Debug("LHFLOW moveToNextLeaderByElection() calling initView(), will increment view to V=%d", currentHV.View()+1)
	currentHV, err := tic.initView(currentHV.View() + 1)
	if err != nil {
		tic.logger.Info("LHFLOW moveToNextLeaderByElection() initView() failed, cannot continue: %s", err)
		return
	}

	newLeader := tic.calcLeaderMemberId(currentHV.View())
	tic.logger.Debug("LHFLOW moveToNextLeaderByElection() calculated newLeader=%s of V=%d", Str(newLeader), currentHV.View())
	var preparedMessages *preparedmessages.PreparedMessages
	if tic.preparedLocally != nil && tic.preparedLocally.isPreparedLocally {
		preparedMessages = preparedmessages.ExtractPreparedMessages(currentHV.Height(), tic.preparedLocally.latestView, tic.storage, tic.QuorumSize)
	}
	vcm := tic.messageFactory.CreateViewChangeMessage(currentHV.Height(), currentHV.View(), preparedMessages)

	if err := tic.isLeader(tic.myMemberId, currentHV.View()); err == nil {
		tic.logger.Debug("LHFLOW moveToNextLeaderByElection() I WILL BE LEADER if I get enough VIEW_CHANGE votes. My leadership of V=%d will time out in %s", currentHV.View(), tic.electionTrigger.CalcTimeout(currentHV.View()))
		tic.storage.StoreViewChange(vcm)
		tic.checkElected(currentHV.Height(), currentHV.View())
	} else {
		tic.logger.Debug("LHFLOW LHMSG SEND VIEW_CHANGE to %s in moveToNextLeader() (I'M NOT LEADER: %s) (msg: H=%d V=%d sender=%s)",
			newLeader, err, vcm.BlockHeight(), vcm.View(), Str(vcm.SenderMemberId()))
		if sendErr := tic.sendConsensusMessageToSpecificMember(newLeader, vcm); sendErr != nil {
			tic.logger.Info("LHMSG SEND VIEW_CHANGE to %s FAILED - %s", newLeader, sendErr)
		}
	}
	if updateMetrics != nil {
		updateMetrics(metrics.NewElectionMetrics(newLeader, currentHV.View()))
	}
}

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

func (tic *TermInCommittee) checkElected(height primitives.BlockHeight, view primitives.View) {
	if tic.latestViewThatProcessedVCMOrNVM >= view {
		tic.logger.Debug("checkElected() already latestViewThatProcessedVCMOrNVM=%d is greater or equal to received view=%d, skipping", tic.latestViewThatProcessedVCMOrNVM, view)
		return
	}
	vcms, ok := tic.storage.GetViewChangeMessages(height, view)
	minimumNodes := tic.QuorumSize
	if !ok {
		tic.logger.Info("checkElected() could not get stored VIEW_CHANGE messages, skipping")
		return
	}

	if len(vcms) < minimumNodes {
		tic.logger.Debug("checkElected() stored %d of %d VIEW_CHANGE messages", len(vcms), minimumNodes)
		return
	}
	tic.logger.Debug("checkElected() stored %d of %d VIEW_CHANGE messages", len(vcms), minimumNodes)
	tic.logger.Debug("checkElected() has enough VIEW_CHANGE messages, proceeding to onElectedByViewChange() with V=%d", view)
	tic.onElectedByViewChange(view, vcms[:minimumNodes])
}

func (tic *TermInCommittee) onElectedByViewChange(view primitives.View, viewChangeMessages []*interfaces.ViewChangeMessage) {
	tic.latestViewThatProcessedVCMOrNVM = view
	tic.logger.Debug("LHFLOW onElectedByViewChange() I AM THE LEADER BY VIEW CHANGE for V=%d, now calling initView()", view)
	currentHeightView, err := tic.initView(view)
	if err != nil {
		tic.logger.Debug("LHFLOW onElectedByViewChange() failed: %s", err)
		return
	}
	block, blockHash := blockextractor.GetLatestBlockFromViewChangeMessages(viewChangeMessages)
	if block == nil {
		tic.logger.Debug("LHFLOW onElectedByViewChange() MISSING BLOCK IN VIEW_CHANGE, calling RequestNewBlockProposal()")

		ctx, err := tic.State.Contexts.For(currentHeightView)
		if err != nil {
			tic.logger.Info("LHFLOW onElectedByViewChange() not sending NEW_VIEW - %e", err)
			return
		}

		block, blockHash = tic.blockUtils.RequestNewBlockProposal(ctx, tic.State.Height(), tic.myMemberId, tic.prevBlock)
		if ctx.Err() != nil {
			tic.logger.Info("LHFLOW onElectedByViewChange() RequestNewBlockProposal() context canceled, not sending NEW_VIEW - %s", ctx.Err())
			return
		}
		tic.logger.Debug("LHFLOW onElectedByViewChange() SEND NEW_VIEW with the new block that was returned from RequestNewBlockProposal()")
	} else {
		tic.logger.Debug("LHFLOW onElectedByViewChange() SEND NEW_VIEW with the block with H=%d from the latest VIEW_CHANGE messages", block.Height())
	}
	ppmContentBuilder := tic.messageFactory.CreatePreprepareMessageContentBuilder(tic.State.Height(), view, block, blockHash)
	ppm := tic.messageFactory.CreatePreprepareMessageFromContentBuilder(ppmContentBuilder, block)
	confirmations := interfaces.ExtractConfirmationsFromViewChangeMessages(viewChangeMessages)
	nvm := tic.messageFactory.CreateNewViewMessage(tic.State.Height(), view, ppmContentBuilder, confirmations, block)
	tic.storage.StorePreprepare(ppm)
	tic.logger.Debug("LHMSG SEND NEW_VIEW (msg: H=%d V=%d sender=%s)",
		nvm.BlockHeight(), nvm.View(), Str(nvm.SenderMemberId()))
	if err := tic.sendConsensusMessage(nvm); err != nil {
		tic.logger.Info("LHMSG SEND NEW_VIEW FAILED - %s", err)
	}
}

func (tic *TermInCommittee) sendConsensusMessage(message interfaces.ConsensusMessage) error {
	tic.logger.Debug("LHMSG SEND sendConsensusMessage() target=ALL, msgType=%v", message.MessageType())
	rawMessage := interfaces.CreateConsensusRawMessage(message)
	return tic.communication.SendConsensusMessage(context.TODO(), tic.otherCommitteeMembersMemberIds, rawMessage)
}

func (tic *TermInCommittee) sendConsensusMessageToSpecificMember(targetMemberId primitives.MemberId, message interfaces.ConsensusMessage) error {
	tic.logger.Debug("LHMSG SEND sendConsensusMessageToSpecificMember() target=%s, msgType=%v", Str(targetMemberId), message.MessageType())
	rawMessage := interfaces.CreateConsensusRawMessage(message)
	return tic.communication.SendConsensusMessage(context.TODO(), []primitives.MemberId{targetMemberId}, rawMessage)
}

func (tic *TermInCommittee) HandlePrePrepare(ppm *interfaces.PreprepareMessage) {
	tic.logger.Debug("LHMSG RECEIVED PREPREPARE (msg: H=%d V=%d sender=%s)",
		ppm.BlockHeight(), ppm.View(), Str(ppm.SenderMemberId()))

	if err := tic.validatePreprepare(ppm); err != nil {
		tic.logger.Info("LHMSG RECEIVED PREPREPARE IGNORE: validatePreprepare() failed: %s", err)
		return
	}

	header := ppm.Content().SignedHeader()

	ctx, err := tic.State.Contexts.For(state.NewHeightView(header.BlockHeight(), header.View()))
	if err != nil {
		tic.logger.Info("LHFLOW LHMSG RECEIVED PREPREPARE IGNORE - %e", err)
		return
	}

	// TODO Is this the correct memberId or should it be ppm.Content().Sender().MemberId ?
	err = tic.blockUtils.ValidateBlockProposal(ctx, ppm.BlockHeight(), tic.calcLeaderMemberId(header.View()), ppm.Block(), ppm.Content().SignedHeader().BlockHash(), tic.prevBlock)
	if err != nil {
		tic.logger.Info("LHMSG RECEIVED PREPREPARE IGNORE: blockUtils.ValidateBlockProposal() failed: %s", err)
		return
	}

	if ctx.Err() != nil { // TODO required?
		tic.logger.Info("LHFLOW HandlePrePrepare() ValidateBlockProposal - %s", ctx.Err())
		return
	}

	tic.processPreprepare(ppm)
}

func (tic *TermInCommittee) validatePreprepare(ppm *interfaces.PreprepareMessage) error {
	blockHeight := ppm.BlockHeight()
	if tic.hasPreprepare(blockHeight, ppm.View()) {
		errMsg := fmt.Sprintf("already stored Preprepare for H=%d V=%d", blockHeight, ppm.View())
		tic.logger.Debug("LHMSG RECEIVED PREPREPARE IGNORE: hasPreprepare(): %s", errMsg)
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

func (tic *TermInCommittee) processPreprepare(ppm *interfaces.PreprepareMessage) {
	header := ppm.Content().SignedHeader()
	if tic.State.View() != header.View() {
		tic.logger.Debug("processPreprepare() message from incorrect view %d", header.View())
		return
	}

	pm := tic.messageFactory.CreatePrepareMessage(header.BlockHeight(), header.View(), header.BlockHash())
	tic.storage.StorePreprepare(ppm)
	tic.storage.StorePrepare(pm)
	tic.logger.Debug("LHMSG SEND PREPARE (msg: H=%d V=%d sender=%s)",
		pm.BlockHeight(), pm.View(), Str(pm.SenderMemberId()))
	if err := tic.sendConsensusMessage(pm); err != nil {
		tic.logger.Info("LHMSG SEND PREPARE FAILED - %s", err)
	}

	if err := tic.checkPreparedLocally(header.BlockHeight(), header.View(), header.BlockHash()); err != nil {
		tic.logger.Debug("checkPreparedLocally: err=%v", err)
	}
}

func (tic *TermInCommittee) HandlePrepare(pm *interfaces.PrepareMessage) {
	tic.logger.Debug("LHMSG RECEIVED PREPARE (msg: H=%d V=%d sender=%s)",
		pm.BlockHeight(), pm.View(), Str(pm.SenderMemberId()))
	header := pm.Content().SignedHeader()
	sender := pm.Content().Sender()

	if err := tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender); err != nil {
		tic.logger.Info("LHMSG RECEIVED PREPARE IGNORE - verification failed for Prepare block-height=%v view=%d block-hash=%s err=%v", header.BlockHeight(), header.View(), header.BlockHash(), err)
		return
	}
	if header.View() < tic.State.View() {
		tic.logger.Debug("LHMSG RECEIVED PREPARE IGNORE - prepare view %v is less than current term's view %v", header.View(), tic.State.View())
		return
	}
	if err := tic.isLeader(sender.MemberId(), header.View()); err == nil {
		tic.logger.Debug("LHMSG RECEIVED PREPARE IGNORE - prepare received from leader (only preprepare can be received from leader)")
		return
	}
	tic.storage.StorePrepare(pm)
	if header.View() > tic.State.View() {
		tic.logger.Debug("LHMSG RECEIVED PREPARE STORE in future view %d", header.View())
	}
	if err := tic.checkPreparedLocally(header.BlockHeight(), header.View(), header.BlockHash()); err != nil {
		tic.logger.Debug("checkPreparedLocally: err=%v", err)
	}
}

func (tic *TermInCommittee) checkPreparedLocally(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) error {
	v, ok := tic.getPreparedLocally()
	if ok && v == view {
		return errors.Errorf("already in PHASE PREPARED for V=%d", view)
	}

	if err := tic.isPreprepared(blockHeight, view, blockHash); err != nil {
		return errors.Wrap(err, "isPreprepared failed")
	}

	countPrepared := tic.countPrepared(blockHeight, view, blockHash)
	isPrepared := countPrepared >= tic.QuorumSize-1
	tic.logger.Debug("LHMSG Check if in PHASE PREPARED: stored %d of %d PREPARE messages", countPrepared, tic.QuorumSize-1)
	if isPrepared {
		tic.onPreparedLocally(blockHeight, view, blockHash)
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

func (tic *TermInCommittee) onPreparedLocally(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	tic.setPreparedLocally(view)
	tic.logger.Debug("LHFLOW LHMSG PHASE PREPARED, PreparedLocally set to V=%d", view)
	cm := tic.messageFactory.CreateCommitMessage(blockHeight, view, blockHash)
	tic.storage.StoreCommit(cm)
	tic.logger.Debug("LHMSG SEND COMMIT (msg: H=%d V=%d sender=%s)",
		cm.BlockHeight(), cm.View(), Str(cm.SenderMemberId()))
	if err := tic.sendConsensusMessage(cm); err != nil {
		tic.logger.Info("LHMSG SEND COMMIT FAILED - %s", err)
	}
	tic.checkCommitted(blockHeight, view, blockHash)
}

func (tic *TermInCommittee) HandleCommit(cm *interfaces.CommitMessage) {
	tic.logger.Debug("LHMSG RECEIVED COMMIT (msg: H=%d V=%d sender=%s)",
		cm.BlockHeight(), cm.View(), Str(cm.SenderMemberId()))
	header := cm.Content().SignedHeader()
	sender := cm.Content().Sender()

	if err := tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender); err != nil {
		tic.logger.Info("LHMSG RECEIVED COMMIT IGNORE - verification failed for Commit block-height=%d view=%d block-hash=%s err=%v", header.BlockHeight(), header.View(), header.BlockHash(), err)
		return
	}
	tic.logger.Debug("LHMSG RECEIVED COMMIT STORE")
	tic.storage.StoreCommit(cm)
	tic.checkCommitted(header.BlockHeight(), header.View(), header.BlockHash())
}

func (tic *TermInCommittee) checkCommitted(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	if tic.committedBlock != nil {
		tic.logger.Debug("LHMSG RECEIVED COMMIT IGNORE - already committed H=%d", tic.committedBlock.Height())
		return
	}
	if err := tic.isPreprepared(blockHeight, view, blockHash); err != nil {
		tic.logger.Debug("LHMSG RECEIVED COMMIT IGNORE - is not preprepared, err=%v", err)
		return
	}
	commits, ok := tic.storage.GetCommitMessages(blockHeight, view, blockHash)
	if !ok || len(commits) < tic.QuorumSize {
		tic.logger.Debug("LHMSG RECEIVED COMMIT - stored %d of %d COMMIT messages", len(commits), tic.QuorumSize)
		return
	}

	tic.logger.Debug("LHMSG RECEIVED COMMIT - stored %d of %d COMMIT messages", len(commits), tic.QuorumSize)

	ppm, ok := tic.storage.GetPreprepareMessage(blockHeight, view)
	if !ok {
		tic.logger.Debug("LHMSG RECEIVED COMMIT IGNORE - missing PPM in Commit message")
		return
	}

	ctx, err := tic.State.Contexts.For(state.NewHeightView(blockHeight, MaxView)) // umbrella context for current term
	if err != nil {
		tic.logger.Debug("LHMSG RECEIVED COMMIT IGNORE - %e", err)
		return
	}

	// --- At this point we are convinced that the block needs to be committed ---
	tic.sendCommitIfNotAlreadySent(commits, blockHeight, view, blockHash)
	tic.committedBlock = ppm.Block()
	tic.logger.Debug("LHFLOW LHMSG PHASE COMMITTED CommittedBlock set to H=%d, calling onCommit() with H=%d V=%d block-hash=%s num-commit-messages=%d",
		ppm.Block().Height(), blockHeight, view, blockHash, len(commits))
	tic.onCommit(ctx, ppm.Block(), commits)
}

func (tic *TermInCommittee) sendCommitIfNotAlreadySent(commits []*interfaces.CommitMessage, blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) {
	var iSentCommitMessage bool
	for _, msg := range commits {
		if msg.SenderMemberId().Equal(tic.myMemberId) {
			iSentCommitMessage = true
			break
		}
	}
	if !iSentCommitMessage {
		cm := tic.messageFactory.CreateCommitMessage(blockHeight, view, blockHash)
		tic.logger.Debug("LHMSG SEND COMMIT [checkCommitted] because I did not send it during onPreparedLocally")
		if err := tic.sendConsensusMessage(cm); err != nil {
			tic.logger.Info("LHMSG SEND COMMIT FAILED [checkCommitted] - %s", err)
		}
	}
}

func (tic *TermInCommittee) HandleViewChange(vcm *interfaces.ViewChangeMessage) {
	tic.logger.Debug("LHMSG RECEIVED VIEW_CHANGE (msg: H=%d V=%d sender=%s)",
		vcm.BlockHeight(), vcm.View(), Str(vcm.SenderMemberId()))

	if err := tic.isViewChangeAccepted(tic.myMemberId, tic.State.View(), vcm.Content()); err != nil {
		tic.logger.Debug("LHMSG RECEIVED VIEW_CHANGE IGNORE - %s", err)
		return
	}

	if err := tic.isViewChangeValid(tic.myMemberId, tic.State.View(), vcm.Content()); err != nil {
		tic.logger.Info("LHMSG RECEIVED VIEW_CHANGE IGNORE - invalid VIEW_CHANGE: %s", err)
		return
	}

	header := vcm.Content().SignedHeader()
	if vcm.Block() != nil && header.PreparedProof() != nil {
		isValidDigest := tic.blockUtils.ValidateBlockCommitment(vcm.BlockHeight(), vcm.Block(), header.PreparedProof().PreprepareBlockRef().BlockHash())
		if !isValidDigest {
			tic.logger.Info("LHMSG RECEIVED VIEW_CHANGE IGNORE - different block hashes for block provided with message, and the block provided by the PPM in the PreparedProof of the message")
			return
		}
	}

	tic.storage.StoreViewChange(vcm)
	tic.checkElected(header.BlockHeight(), header.View())
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

	if !proofsvalidator.ValidatePreparedProof(tic.State.Height(), vcmView, preparedProof, tic.QuorumSize, tic.keyManager, tic.committeeMembersMemberIds, func(view primitives.View) primitives.MemberId { return tic.calcLeaderMemberId(view) }) {
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

func (tic *TermInCommittee) HandleNewView(nvm *interfaces.NewViewMessage) {
	tic.logger.Debug("LHMSG RECEIVED NEW_VIEW (msg: H=%d V=%d sender=%s)",
		nvm.BlockHeight(), nvm.View(), Str(nvm.SenderMemberId()))
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
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", ignored because the signature verification failed` });
		tic.logger.Info("LHMSG RECEIVED NEW_VIEW IGNORE - keyManager.VerifyConsensusMessage() failed: %s", err)
		return
	}

	calculatedLeaderFromNewView := tic.calcLeaderMemberId(nvmHeader.View())
	if err := tic.isLeader(nvmSender.MemberId(), nvmHeader.View()); err != nil {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], handleNewViewMessage from "${senderId}", rejected because it match the new id (${view})` });
		tic.logger.Info("LHMSG RECEIVED NEW_VIEW IGNORE - sender %s no match for future leader: %s", nvmSender.MemberId(), err)
		return
	}

	if err := tic.validateViewChangeVotes(nvmHeader.BlockHeight(), nvmHeader.View(), viewChangeConfirmations); err != nil {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", votes is invalid` });
		tic.logger.Info("LHMSG RECEIVED NEW_VIEW IGNORE - validateViewChangeVotes failed: %s", err)
		return
	}

	if tic.State.View() > nvmHeader.View() {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", view is from the past` });
		tic.logger.Info("LHMSG RECEIVED NEW_VIEW IGNORE - current view %d is higher than message view %d", tic.State.View(), nvmHeader.View())
		return
	}

	ppmView := ppMessageContent.SignedHeader().View()
	if !ppmView.Equal(nvmHeader.View()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", view doesn't match PP.view` });
		tic.logger.Info("LHMSG RECEIVED NEW_VIEW IGNORE - NewView.view %d and NewView.Preprepare.view %d do not match",
			nvmHeader.View(), ppmView)
		return
	}

	if !ppMessageContent.SignedHeader().BlockHeight().Equal(nvmHeader.BlockHeight()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", blockHeight doesn't match PP.Block()Height` });
		tic.logger.Info("LHMSG RECEIVED NEW_VIEW IGNORE - NewView.BlockHeight and NewView.Preprepare.BlockHeight do not match")
		return
	}

	latestVote := tic.latestViewChangeVote(viewChangeConfirmations)
	if latestVote != nil {

		calculatedLeaderFromViewChange := tic.calcLeaderMemberId(latestVote.SignedHeader().View())
		if !calculatedLeaderFromNewView.Equal(calculatedLeaderFromViewChange) {
			tic.logger.Debug("LHMSG RECEIVED NEW_VIEW IGNORE - calculatedLeaderFromNewView=%s is not the calculated leader %s who should collect these messages", Str(calculatedLeaderFromNewView), Str(calculatedLeaderFromViewChange))
			return
		}

		if err := tic.isViewChangeValid(calculatedLeaderFromNewView, nvmHeader.View(), latestVote); err != nil {
			tic.logger.Info("LHMSG RECEIVED NEW_VIEW IGNORE - NewView.ViewChangeConfirmation (with latest view) is invalid: %s", err)
			return
		}

		// rewrite this mess
		latestVoteBlockHash := latestVote.SignedHeader().PreparedProof().PreprepareBlockRef().BlockHash()
		if latestVoteBlockHash != nil {
			isValidDigest := tic.blockUtils.ValidateBlockCommitment(nvmHeader.BlockHeight(), nvm.Block(), latestVoteBlockHash)
			if !isValidDigest {
				//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", the given _Block (PP._Block) doesn't match the best _Block from the VCProof` });
				tic.logger.Info("LHMSG RECEIVED NEW_VIEW IGNORE - NewView.ViewChangeConfirmation (with latest view) is invalid")
				return
			}
		}
	}

	ppm := interfaces.NewPreprepareMessage(ppMessageContent, nvm.Block())

	// leader proposed a new block in this view, checking its proposal
	if latestVote == nil {
		header := ppm.Content().SignedHeader()

		ctx, err := tic.State.Contexts.For(state.NewHeightView(nvmHeader.BlockHeight(), nvm.View()))
		if err != nil {
			tic.logger.Info("LHFLOW LHMSG RECEIVED NEW_VIEW IGNORE - %e", err)
			return
		}

		// TODO Is this the correct member Id or should it be ppm.Content().Sender().MemberId()?
		err = tic.blockUtils.ValidateBlockProposal(ctx, ppm.BlockHeight(), tic.calcLeaderMemberId(header.View()), ppm.Block(), ppm.Content().SignedHeader().BlockHash(), tic.prevBlock)
		if err != nil {
			tic.logger.Info("LHFLOW LHMSG RECEIVED NEW_VIEW IGNORE - Proposed block failed ValidateBlockProposal: %s", err)
			return
		}

		if ctx.Err() != nil { // TODO required?
			tic.logger.Info("LHFLOW LHMSG RECEIVED NEW_VIEW IGNORE - ValidateBlockProposal - %s", ctx.Err())
			return
		}
	}

	if err := tic.validatePreprepare(ppm); err == nil {
		tic.latestViewThatProcessedVCMOrNVM = nvmHeader.View()
		tic.logger.Debug("LHFLOW LHMSG RECEIVED NEW_VIEW OK - calling initView(). latestViewThatProcessedVCMOrNVM set to V=%d", tic.latestViewThatProcessedVCMOrNVM)
		if _, err := tic.initView(nvmHeader.View()); err != nil {
			tic.logger.Debug("LHFLOW LHMSG HandleNewView() - initView() failed: %s", err)
			return
		}
		tic.processPreprepare(ppm)
	} else {
		tic.logger.Info("LHFLOW LHMSG RECEIVED NEW_VIEW FAILED validation of PPM: %s", err)
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

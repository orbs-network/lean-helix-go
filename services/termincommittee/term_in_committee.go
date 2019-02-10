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

func Str(memberId primitives.MemberId) string {
	if memberId == nil {
		return ""
	}
	return memberId.String()[:6]
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

	result.startTerm(ctx)
	return result
}

func ToCommitteeMembersStr(members []primitives.MemberId) string {

	strs := make([]string, 1)
	for _, member := range members {
		strs = append(strs, Str(member))
	}
	return strings.Join(strs, ",")
}

func panicOnLessThanMinimumCommitteeMembers(committeeMembers []primitives.MemberId) {
	if len(committeeMembers) < LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS {
		panic(fmt.Sprintf("LH Received only %d committee members, but the hard minimum is %d", len(committeeMembers), LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS))
	}
}

func (tic *TermInCommittee) startTerm(ctx context.Context) {
	tic.setNotPreparedLocally()
	tic.initView(ctx, 0)
	if tic.isLeader(tic.myMemberId, tic.view) {
		tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW startTerm() I AM THE LEADER OF FIRST VIEW, requesting new block")
		block, blockHash := tic.blockUtils.RequestNewBlockProposal(ctx, tic.height, tic.prevBlock)
		ppm := tic.messageFactory.CreatePreprepareMessage(tic.height, tic.view, block, blockHash)

		tic.storage.StorePreprepare(ppm)
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG SEND PREPREPARE")
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
	tic.storage.ClearBlockHeightLogs(tic.height)
}

func (tic *TermInCommittee) calcLeaderMemberId(view primitives.View) primitives.MemberId {
	index := int(view) % len(tic.committeeMembersMemberIds)
	return tic.committeeMembersMemberIds[index]
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
	if tic.isLeader(tic.myMemberId, tic.view) {
		tic.logger.Info(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW moveToNextLeader() I AM THE LEADER BY VIEW CHANGE. My leadership will time out in %s", tic.electionTrigger.CalcTimeout(view))
		tic.storage.StoreViewChange(vcm)
		tic.checkElected(ctx, tic.height, tic.view)
	} else {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG SEND VIEW_CHANGE (I'm not leader) moveToNextLeader()")
		tic.sendConsensusMessageToSpecificMember(ctx, newLeader, vcm)
	}
	if onElectionCB != nil {
		onElectionCB(metrics.NewElectionMetrics(newLeader, tic.view))
	}
}

func (tic *TermInCommittee) isLeader(memberId primitives.MemberId, v primitives.View) bool {
	return memberId.Equal(tic.calcLeaderMemberId(v))
}

// TODO v1 breakdown to separate if's and log each
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
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "checkElected() stored %d of %d VIEW_CHANGE messages", len(vcms), minimumNodes)
		return
	}
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "checkElected() stored %d of %d VIEW_CHANGE messages", len(vcms), minimumNodes)
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "checkElected() has enough VIEW_CHANGE messages, proceeding to onElected()", view)
	tic.onElected(ctx, view, vcms[:minimumNodes])
}

func (tic *TermInCommittee) onElected(ctx context.Context, view primitives.View, viewChangeMessages []*interfaces.ViewChangeMessage) {
	tic.newViewLocally = view
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW I AM ELECTED for NewViewLocally=%d (onElected), now calling SetView()", tic.newViewLocally)
	tic.SetView(ctx, view)
	block, blockHash := blockextractor.GetLatestBlockFromViewChangeMessages(viewChangeMessages)
	if block == nil {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW onElected() MISSING BLOCK IN VIEW_CHANGE, requesting new block")
		block, blockHash = tic.blockUtils.RequestNewBlockProposal(ctx, tic.height, tic.prevBlock)
	}
	ppmContentBuilder := tic.messageFactory.CreatePreprepareMessageContentBuilder(tic.height, view, block, blockHash)
	ppm := tic.messageFactory.CreatePreprepareMessageFromContentBuilder(ppmContentBuilder, block)
	confirmations := interfaces.ExtractConfirmationsFromViewChangeMessages(viewChangeMessages)
	nvm := tic.messageFactory.CreateNewViewMessage(tic.height, view, ppmContentBuilder, confirmations, block)
	tic.storage.StorePreprepare(ppm)
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG SEND NEW_VIEW")
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
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPREPARE IGNORE - err=%v", err)
		return
	}

	err := tic.blockUtils.ValidateBlockProposal(ctx, ppm.BlockHeight(), ppm.Block(), ppm.Content().SignedHeader().BlockHash(), tic.prevBlock)
	if err != nil {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPREPARE IGNORE - err=%v", err)
	}

	tic.processPreprepare(ctx, ppm)

}

func (tic *TermInCommittee) validatePreprepare(ctx context.Context, ppm *interfaces.PreprepareMessage) error {
	blockHeight := ppm.BlockHeight()
	if tic.hasPreprepare(blockHeight, ppm.View()) {
		errMsg := fmt.Sprintf("already stored Preprepare for H=%d V=%d", blockHeight, ppm.View())
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED PREPREPARE IGNORE - %s", errMsg)
		return errors.New(errMsg)
	}

	header := ppm.Content().SignedHeader()
	sender := ppm.Content().Sender()
	if err := tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender); err != nil {
		return errors.Wrapf(err, "verification failed for sender %s signature on header", Str(sender.MemberId()))
	}

	if !tic.isLeader(sender.MemberId(), ppm.View()) {
		// Log
		return fmt.Errorf("sender %s is not leader. ExpectedLeader=%s", Str(sender.MemberId()), Str(tic.calcLeaderMemberId(ppm.View())))
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
	if tic.isLeader(sender.MemberId(), header.View()) {
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
	tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED VIEW_CHANGE (H=%d V=%d sender=%s)", vcm.BlockHeight(), vcm.View(), Str(vcm.SenderMemberId()))
	if !tic.isViewChangeValid(tic.myMemberId, tic.view, vcm.Content()) { // errors logged inside this func
		return
	}

	header := vcm.Content().SignedHeader()
	if vcm.Block() != nil && header.PreparedProof() != nil {
		isValidDigest := tic.blockUtils.ValidateBlockCommitment(vcm.BlockHeight(), vcm.Block(), header.PreparedProof().PreprepareBlockRef().BlockHash())
		if !isValidDigest {
			tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED VIEW_CHANGE IGNORE - different block hashes for block provided with message, and the block provided by the PPM in the PreparedProof of the message")
			return
		}
	}

	tic.storage.StoreViewChange(vcm)
	tic.checkElected(ctx, header.BlockHeight(), header.View())
}

// TODO change to return error
func (tic *TermInCommittee) isViewChangeValid(expectedLeaderFromNewView primitives.MemberId, currentView primitives.View, vcm *protocol.ViewChangeMessageContent) bool {
	header := vcm.SignedHeader()
	sender := vcm.Sender()
	vcmView := header.View()
	preparedProof := header.PreparedProof()

	if err := tic.keyManager.VerifyConsensusMessage(header.BlockHeight(), header.Raw(), sender); err != nil {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED VIEW_CHANGE IGNORE - VerifyConsensusMessage() failed. err=%v", err)
		return false
	}

	if currentView > vcmView {
		tic.logger.Debug(L.LC(tic.height, currentView, tic.myMemberId), "LHMSG RECEIVED VIEW_CHANGE IGNORE - message view %s is older than current term's view %s", vcmView, currentView)
		return false
	}

	if !proofsvalidator.ValidatePreparedProof(tic.height, vcmView, preparedProof, tic.QuorumSize, tic.keyManager, tic.committeeMembersMemberIds, func(view primitives.View) primitives.MemberId { return tic.calcLeaderMemberId(view) }) {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED VIEW_CHANGE IGNORE - failed ValidatePreparedProof()")
		return false
	}

	calculatedLeaderFromViewChange := tic.calcLeaderMemberId(vcmView)
	if !expectedLeaderFromNewView.Equal(calculatedLeaderFromViewChange) {
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED VIEW_CHANGE IGNORE - targetLeaderMemberId=%s different from calculatedLeaderFromViewChange=%s", Str(expectedLeaderFromNewView), Str(calculatedLeaderFromViewChange))
		return false
	}

	return true

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
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", ignored because the signature verification failed` });
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - verify failed. err=%v", err)
		return
	}

	calculatedLeaderFromNewView := tic.calcLeaderMemberId(nvmHeader.View())
	if !tic.isLeader(nvmSender.MemberId(), nvmHeader.View()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", rejected because it match the new id (${view})` });
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - no match for future leader")
		return
	}

	if !tic.validateViewChangeVotes(nvmHeader.BlockHeight(), nvmHeader.View(), viewChangeConfirmations) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", votes is invalid` });
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - validateViewChangeVotes failed")
		return
	}

	if tic.view > nvmHeader.View() {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", view is from the past` });
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - current view is higher than message view")
		return
	}

	if !ppMessageContent.SignedHeader().View().Equal(nvmHeader.View()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", view doesn't match PP.view` });
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - NewView.view and NewView.Preprepare.view do not match")
		return
	}

	if !ppMessageContent.SignedHeader().BlockHeight().Equal(nvmHeader.BlockHeight()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", blockHeight doesn't match PP.Block()Height` });
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - NewView.BlockHeight and NewView.Preprepare.BlockHeight do not match")
		return
	}

	latestVote := tic.latestViewChangeVote(viewChangeConfirmations)
	if latestVote != nil {
		viewChangeMessageValid := tic.isViewChangeValid(calculatedLeaderFromNewView, nvmHeader.View(), latestVote)
		if !viewChangeMessageValid {
			//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", view change votes are invalid` });
			tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - NewView.ViewChangeConfirmation (with latest view) is invalid")
			return
		}

		// rewrite this mess
		latestVoteBlockHash := latestVote.SignedHeader().PreparedProof().PreprepareBlockRef().BlockHash()
		if latestVoteBlockHash != nil {
			isValidDigest := tic.blockUtils.ValidateBlockCommitment(nvmHeader.BlockHeight(), nvm.Block(), latestVoteBlockHash)
			if !isValidDigest {
				//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], HandleNewView from "${senderId}", the given _Block (PP._Block) doesn't match the best _Block from the VCProof` });
				tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHMSG RECEIVED NEW_VIEW IGNORE - NewView.ViewChangeConfirmation (with latest view) is invalid")
				return
			}
		}
	}

	ppm := interfaces.NewPreprepareMessage(ppMessageContent, nvm.Block())

	// leader proposed a new block in this view, checking its proposal
	if latestVote == nil {
		err := tic.blockUtils.ValidateBlockProposal(ctx, ppm.BlockHeight(), ppm.Block(), ppm.Content().SignedHeader().BlockHash(), tic.prevBlock)
		if err != nil {
			tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "Proposed block failed ValidateBlockProposal - err=%v", err)
			return
		}
	}

	if err := tic.validatePreprepare(ctx, ppm); err == nil {
		tic.newViewLocally = nvmHeader.View()
		tic.logger.Debug(L.LC(tic.height, tic.view, tic.myMemberId), "LHFLOW NewViewLocally set to V=%d (HandleNewView), calling SetView()", tic.newViewLocally)
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

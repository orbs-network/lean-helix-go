package leanhelix

import (
	"context"
	"fmt"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"math"
	"sort"
)

type LeanHelixTerm struct {
	ctx context.Context
	KeyManager
	NetworkCommunication
	Storage
	electionTrigger ElectionTrigger
	BlockUtils
	messageFactory                  *MessageFactory
	filter                          *ConsensusMessageFilter
	myPublicKey                     Ed25519PublicKey
	committeeMembersPublicKeys      []Ed25519PublicKey
	otherCommitteeMembersPublicKeys []Ed25519PublicKey
	height                          BlockHeight
	view                            View
	preparedLocally                 bool
	committedBlock                  Block
	leaderPublicKey                 Ed25519PublicKey
	newViewLocally                  View
	logger                          Logger
}

func NewLeanHelixTerm(ctx context.Context, config *Config, filter *ConsensusMessageFilter, newBlockHeight BlockHeight) *LeanHelixTerm {
	keyManager := config.KeyManager
	blockUtils := config.BlockUtils
	myPK := keyManager.MyPublicKey()
	comm := config.NetworkCommunication
	messageFactory := NewMessageFactory(keyManager)

	// TODO Implement me!
	randomSeed := uint64(12345)
	// TODO Implement me!
	maxCommitteeSize := uint32(4)
	committeeMembers := comm.RequestOrderedCommittee(ctx, newBlockHeight, randomSeed, maxCommitteeSize)

	panicOnLessThanMinimumCommitteeMembers(config.OverrideMinimumCommitteeMembers, committeeMembers)

	otherCommitteeMembers := make([]Ed25519PublicKey, 0)
	for _, member := range committeeMembers {
		if !member.Equal(myPK) {
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
		height:                          newBlockHeight,
		KeyManager:                      keyManager,
		NetworkCommunication:            comm,
		Storage:                         config.Storage,
		electionTrigger:                 config.ElectionTrigger,
		BlockUtils:                      blockUtils,
		committeeMembersPublicKeys:      committeeMembers,
		otherCommitteeMembersPublicKeys: otherCommitteeMembers,
		messageFactory:                  messageFactory,
		myPublicKey:                     myPK,
		filter:                          filter,
		logger:                          config.Logger,
	}

	newTerm.logger.Debug("H %d V 0 NewLeanHelixTerm: myID=%s committeeMembersCount=%d", newBlockHeight, keyManager.MyPublicKey()[:3], len(committeeMembers))
	return newTerm
}

func panicOnLessThanMinimumCommitteeMembers(minimum int, committeeMembers []Ed25519PublicKey) {

	if minimum == 0 {
		minimum = LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS
	}

	if len(committeeMembers) < minimum {
		panic(fmt.Sprintf("LH Received only %d committee members, but the hard minimum is %d", len(committeeMembers), LEAN_HELIX_HARD_MINIMUM_COMMITTEE_MEMBERS))
	}
}

func (term *LeanHelixTerm) WaitForBlock(ctx context.Context) Block {
	term.startTerm(ctx)

	for {
		ctxWithElectionTrigger := term.electionTrigger.CreateElectionContextForView(ctx, term.view)
		message, err := term.filter.WaitForMessage(ctxWithElectionTrigger, term.height)

		if err != nil {
			term.logger.Debug("H %s V ? WaitForBlock() err=%v", term.height, err)
			if ctx.Err() == nil {
				term.moveToNextLeader(ctx)
				continue
			}
			return nil
		}

		term.handleMessage(ctx, message)
		if term.committedBlock != nil {
			return term.committedBlock
		}
	}
	return nil
}

func (term *LeanHelixTerm) handleMessage(ctx context.Context, consensusMessage ConsensusMessage) {
	switch message := consensusMessage.(type) {
	case *PreprepareMessage:
		term.onReceivePreprepare(ctx, message)
	case *PrepareMessage:
		term.onReceivePrepare(ctx, message)
	case *CommitMessage:
		term.onReceiveCommit(ctx, message)
	case *ViewChangeMessage:
		term.onReceiveViewChange(ctx, message)
	case *NewViewMessage:
		term.onReceiveNewView(ctx, message)
	default:
		panic(fmt.Sprintf("unknown message type: %T", consensusMessage))
	}
}

func (term *LeanHelixTerm) startTerm(ctx context.Context) {

	term.initView(0)
	isLeader := term.IsLeader()
	term.logger.Debug("H %d V 0 startTerm(): leader? %t", term.height, isLeader)
	if isLeader {
		block := term.BlockUtils.RequestNewBlock(ctx, term.height)
		blockHash := term.BlockUtils.CalculateBlockHash(block)
		ppm := term.messageFactory.CreatePreprepareMessage(term.height, term.view, block, blockHash)

		term.Storage.StorePreprepare(ppm)
		term.sendPreprepare(ctx, ppm)
	}
}

func (term *LeanHelixTerm) GetView() View {
	return term.view
}

func (term *LeanHelixTerm) SetView(view View) {
	if term.view != view {
		term.initView(view)
	}
}

func (term *LeanHelixTerm) initView(view View) {
	term.preparedLocally = false
	term.view = view
	term.leaderPublicKey = term.calcLeaderPublicKey(view)
	term.logger.Debug("H %d V %d initView() set leader to %s", term.height, term.view, term.leaderPublicKey[:3])
}

func (term *LeanHelixTerm) Dispose() {
	term.Storage.ClearBlockHeightLogs(term.height)
}

func (term *LeanHelixTerm) calcLeaderPublicKey(view View) Ed25519PublicKey {
	index := int(view) % len(term.committeeMembersPublicKeys)
	return term.committeeMembersPublicKeys[index]
}

func (term *LeanHelixTerm) moveToNextLeader(ctx context.Context) {
	term.SetView(term.view + 1)
	term.logger.Debug("H %d V %d moveToNextLeader() newLeader=%s", term.height, term.view, term.leaderPublicKey[:3])
	preparedMessages := ExtractPreparedMessages(term.height, term.Storage, term.QuorumSize())
	vcm := term.messageFactory.CreateViewChangeMessage(term.height, term.view, preparedMessages)
	term.Storage.StoreViewChange(vcm)
	if term.IsLeader() {
		term.checkElected(ctx, term.height, term.view)
	} else {
		term.sendViewChange(ctx, vcm)
	}
}

func (term *LeanHelixTerm) sendPreprepare(ctx context.Context, message *PreprepareMessage) {
	term.logger.Debug("H %d V %d sendPreprepare()", term.height, term.view)
	rawMessage := message.ToConsensusRawMessage()
	term.NetworkCommunication.SendMessage(ctx, term.otherCommitteeMembersPublicKeys, rawMessage)
}

func (term *LeanHelixTerm) sendPrepare(ctx context.Context, message *PrepareMessage) {
	term.logger.Debug("H %s V %s sendPrepare()", term.height, term.view)
	rawMessage := message.ToConsensusRawMessage()
	term.NetworkCommunication.SendMessage(ctx, term.otherCommitteeMembersPublicKeys, rawMessage)
}

func (term *LeanHelixTerm) sendCommit(ctx context.Context, message *CommitMessage) {
	term.logger.Debug("H %s V %s sendCommit()", term.height, term.view)
	rawMessage := message.ToConsensusRawMessage()
	term.NetworkCommunication.SendMessage(ctx, term.otherCommitteeMembersPublicKeys, rawMessage)
}

func (term *LeanHelixTerm) sendViewChange(ctx context.Context, message *ViewChangeMessage) {
	term.logger.Debug("H %s V %s sendViewChange()", term.height, term.view)
	rawMessage := message.ToConsensusRawMessage()
	term.NetworkCommunication.SendMessage(ctx, []Ed25519PublicKey{term.leaderPublicKey}, rawMessage)
}

func (term *LeanHelixTerm) sendNewView(ctx context.Context, message *NewViewMessage) {
	term.logger.Debug("H %s V %s sendNewView()", term.height, term.view)
	rawMessage := message.ToConsensusRawMessage()
	term.NetworkCommunication.SendMessage(ctx, term.otherCommitteeMembersPublicKeys, rawMessage)
}

func (term *LeanHelixTerm) onReceivePreprepare(ctx context.Context, ppm *PreprepareMessage) {
	term.logger.Debug("H %s V %s onReceivePreprepare()", term.height, term.view)
	if err := term.validatePreprepare(ppm); err != nil {
		term.logger.Debug("H %s V %s onReceivePreprepare() err=%v", err)
	} else {
		term.processPreprepare(ctx, ppm)
	}
}

func (term *LeanHelixTerm) processPreprepare(ctx context.Context, ppm *PreprepareMessage) {
	header := ppm.content.SignedHeader()
	if term.view != header.View() {
		term.logger.Debug("H %s V %s onReceivePreprepare() message from incorrect view %d", term.height, term.view, header.View())
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], processPrePrepare, view doesn't match` });
		return
	}

	pm := term.messageFactory.CreatePrepareMessage(header.BlockHeight(), header.View(), header.BlockHash())
	term.Storage.StorePreprepare(ppm)
	term.Storage.StorePrepare(pm)
	term.sendPrepare(ctx, pm)
	term.checkPrepared(ctx, header.BlockHeight(), header.View(), header.BlockHash())
}

func (term *LeanHelixTerm) validatePreprepare(ppm *PreprepareMessage) error {
	blockHeight := ppm.BlockHeight()
	view := ppm.View()
	if term.hasPreprepare(blockHeight, view) {
		return fmt.Errorf("already received Preprepare for H %s V %s", blockHeight, view)
	}

	header := ppm.Content().SignedHeader()
	sender := ppm.Content().Sender()
	if !term.KeyManager.Verify(header.Raw(), sender) {
		return fmt.Errorf("verification failed for sender %s signature on header", sender.SenderPublicKey()[:3])
	}

	leaderPublicKey := term.calcLeaderPublicKey(view)
	senderPublicKey := sender.SenderPublicKey()
	if !senderPublicKey.Equal(leaderPublicKey) {
		// Log
		return fmt.Errorf("sender %s is not leader", senderPublicKey[:3])
	}

	givenBlockHash := term.BlockUtils.CalculateBlockHash(ppm.Block())
	if !ppm.Content().SignedHeader().BlockHash().Equal(givenBlockHash) {
		return fmt.Errorf("block hash in block and in header are different")
	}

	isValidBlock := term.BlockUtils.ValidateBlock(ppm.Block())

	if !isValidBlock {
		return fmt.Errorf("block validation failed")
	}

	return nil
}

func (term *LeanHelixTerm) hasPreprepare(blockHeight BlockHeight, view View) bool {
	_, ok := term.GetPreprepareMessage(blockHeight, view)
	return ok
}

func (term *LeanHelixTerm) onReceivePrepare(ctx context.Context, pm *PrepareMessage) {
	term.logger.Debug("H %s V %s onReceivePrepare()", pm.BlockHeight(), pm.View())
	header := pm.content.SignedHeader()
	sender := pm.content.Sender()

	if !term.KeyManager.Verify(header.Raw(), sender) {
		fmt.Printf("verification failed for Prepare blockHeight=%v view=%v blockHash=%v\n", header.BlockHeight(), header.View(), header.BlockHash())
		return
	}
	if term.view > header.View() {
		fmt.Printf("prepare view %v is less than OneHeight's view %v\n", header.View(), term.view)
		return
	}
	if term.leaderPublicKey.Equal(sender.SenderPublicKey()) {
		fmt.Printf("prepare received from leader (only preprepare can be received from leader)\n")
		return
	}
	term.Storage.StorePrepare(pm)
	if term.view == header.View() {
		term.checkPrepared(ctx, header.BlockHeight(), header.View(), header.BlockHash())
	}
}

func (term *LeanHelixTerm) onReceiveViewChange(ctx context.Context, vcm *ViewChangeMessage) {
	term.logger.Debug("H %s V %s onReceiveViewChange()", term.height, term.view)
	header := vcm.content.SignedHeader()
	if !term.isViewChangeValid(term.myPublicKey, term.view, vcm.content) {
		fmt.Printf("message ViewChange is not valid\n")
		return
	}
	if vcm.block == nil || header.PreparedProof() == nil {
		fmt.Printf("message ViewChange - block or prepared proof are nil\n")
		return
	}
	calculatedBlockHash := term.BlockUtils.CalculateBlockHash(vcm.block)
	isValidDigest := calculatedBlockHash.Equal(header.PreparedProof().PreprepareBlockRef().BlockHash())
	if !isValidDigest {
		fmt.Printf("different block hashes for block provided with message, and the block provided by the PPM in the PreparedProof of the message\n")
		return
	}
	term.Storage.StoreViewChange(vcm)
	term.checkElected(ctx, header.BlockHeight(), header.View())
}

func (term *LeanHelixTerm) isViewChangeValid(targetLeaderPublicKey Ed25519PublicKey, view View, confirmation *ViewChangeMessageContent) bool {
	header := confirmation.SignedHeader()
	sender := confirmation.Sender()
	newView := header.View()
	preparedProof := header.PreparedProof()

	if !term.KeyManager.Verify(header.Raw(), sender) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${newView}], onReceiveViewChange from "${senderPk}", ignored because the signature verification failed` });
		return false
	}

	if view > newView {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${newView}], onReceiveViewChange from "${senderPk}", ignored because of unrelated view` });
		return false
	}

	if !ValidatePreparedProof(term.height, newView, preparedProof, term.QuorumSize(), term.KeyManager, term.committeeMembersPublicKeys, func(view View) Ed25519PublicKey { return term.calcLeaderPublicKey(view) }) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${newView}], onReceiveViewChange from "${senderPk}", ignored because the preparedProof is invalid` });
		return false
	}

	futureLeaderPublicKey := term.calcLeaderPublicKey(newView)
	if !targetLeaderPublicKey.Equal(futureLeaderPublicKey) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], newView:[${newView}], onReceiveViewChange from "${senderPk}", ignored because the newView doesn't match the target leader` });
		return false
	}

	return true

}

func (term *LeanHelixTerm) checkElected(ctx context.Context, height BlockHeight, view View) {
	if term.newViewLocally < view {
		vcms := term.Storage.GetViewChangeMessages(height, view)
		minimumNodes := term.QuorumSize()
		if len(vcms) >= minimumNodes {
			term.onElected(ctx, view, vcms[:minimumNodes])
		}
	}
}

func (term *LeanHelixTerm) onElected(ctx context.Context, view View, viewChangeMessages []*ViewChangeMessage) {
	term.newViewLocally = view
	term.SetView(view)
	block := GetLatestBlockFromViewChangeMessages(viewChangeMessages)
	if block == nil {
		block = term.BlockUtils.RequestNewBlock(term.ctx, term.height)
	}
	ppmContentBuilder := term.messageFactory.CreatePreprepareMessageContentBuilder(term.height, view, block, term.BlockUtils.CalculateBlockHash(block))
	ppm := term.messageFactory.CreatePreprepareMessageFromContentBuilder(ppmContentBuilder, block)
	confirmations := extractConfirmationsFromViewChangeMessages(viewChangeMessages)
	nvm := term.messageFactory.CreateNewViewMessage(term.height, view, ppmContentBuilder, confirmations, block)
	term.Storage.StorePreprepare(ppm)
	term.sendNewView(ctx, nvm)
}

func (term *LeanHelixTerm) checkPrepared(ctx context.Context, blockHeight BlockHeight, view View, blockHash Uint256) {
	if term.preparedLocally == false {
		if term.isPreprepared(blockHeight, view, blockHash) {
			countPrepared := term.countPrepared(blockHeight, view, blockHash)
			if countPrepared >= term.QuorumSize()-1 {
				term.onPrepared(ctx, blockHeight, view, blockHash)
			}
		}
	}
}

func (term *LeanHelixTerm) onPrepared(ctx context.Context, blockHeight BlockHeight, view View, blockHash Uint256) {
	term.preparedLocally = true
	cm := term.messageFactory.CreateCommitMessage(blockHeight, view, blockHash)
	term.Storage.StoreCommit(cm)
	term.sendCommit(ctx, cm)
	term.checkCommitted(blockHeight, view, blockHash)
}

func (term *LeanHelixTerm) onReceiveCommit(ctx context.Context, cm *CommitMessage) {
	term.logger.Debug("H %s V %s onReceiveCommit()", term.height, term.view)
	header := cm.content.SignedHeader()
	sender := cm.content.Sender()

	if !term.KeyManager.Verify(header.Raw(), sender) {
		fmt.Printf("verification failed for Commit blockHeight=%v view=%v blockHash=%v\n", header.BlockHeight(), header.View(), header.BlockHash())
		return
	}
	term.Storage.StoreCommit(cm)
	if term.view == header.View() {
		term.checkCommitted(header.BlockHeight(), header.View(), header.BlockHash())
	}
}

func (term *LeanHelixTerm) checkCommitted(blockHeight BlockHeight, view View, blockHash Uint256) {
	term.logger.Debug("H %s V %s checkCommitted() H %s V %s BlockHash %s ", term.height, term.view, blockHeight, view, blockHash)
	if term.committedBlock != nil {
		return
	}
	if !term.isPreprepared(blockHeight, view, blockHash) {
		return
	}
	commits := term.Storage.GetCommitSendersPKs(blockHeight, view, blockHash)
	if len(commits) < term.QuorumSize() {
		return
	}
	ppm, ok := term.Storage.GetPreprepareMessage(blockHeight, view)
	if !ok {
		// log
		term.logger.Info("H %s V %s checkCommitted() missing PPM", )
		return
	}
	term.logger.Info("H %s V %s checkCommitted() COMMITTED H %s V %s BlockHash %s ", term.height, term.view, blockHeight, view, blockHash)
	term.committedBlock = ppm.block
}

func (term *LeanHelixTerm) validateViewChangeVotes(targetBlockHeight BlockHeight, targetView View, confirmations []*ViewChangeMessageContent) bool {
	if len(confirmations) < term.QuorumSize() {
		return false
	}

	set := make(map[string]bool)

	// Verify that all _Block heights and views match, and all public keys are unique
	// TODO consider refactor here, the purpose of this code is not apparent
	for _, confirmation := range confirmations {
		senderPublicKeyStr := string(confirmation.Sender().SenderPublicKey())
		if confirmation.SignedHeader().BlockHeight() != targetBlockHeight {
			return false
		}
		if confirmation.SignedHeader().View() != targetView {
			return false
		}
		if set[senderPublicKeyStr] {
			return false
		}
		set[senderPublicKeyStr] = true
	}

	return true

}

func (term *LeanHelixTerm) latestViewChangeVote(confirmations []*ViewChangeMessageContent) *ViewChangeMessageContent {
	res := make([]*ViewChangeMessageContent, 0, len(confirmations))
	for _, confirmation := range confirmations {
		if confirmation.SignedHeader().PreparedProof() != nil {
			res = append(res, confirmation)
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[j].SignedHeader().PreparedProof().PreprepareBlockRef().View() > res[i].SignedHeader().PreparedProof().PreprepareBlockRef().View()
	})

	if len(res) > 0 {
		return res[0]
	} else {
		return nil
	}
}

func (term *LeanHelixTerm) onReceiveNewView(ctx context.Context, nvm *NewViewMessage) {
	term.logger.Debug("H %s V %s onReceiveNewView()", term.height, term.view)
	header := nvm.Content().SignedHeader()
	sender := nvm.Content().Sender()
	ppMessageContent := nvm.Content().PreprepareMessageContent()
	viewChangeConfirmationsIter := header.ViewChangeConfirmationsIterator()
	viewChangeConfirmations := make([]*ViewChangeMessageContent, 0, 1)
	for {
		if !viewChangeConfirmationsIter.HasNext() {
			break
		}
		viewChangeConfirmations = append(viewChangeConfirmations, viewChangeConfirmationsIter.NextViewChangeConfirmations())
	}

	if !term.KeyManager.Verify(header.Raw(), sender) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", ignored because the signature verification failed` });
		fmt.Printf("verify failed\n")
		return
	}

	futureLeaderId := term.calcLeaderPublicKey(header.View())
	if !sender.SenderPublicKey().Equal(futureLeaderId) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", rejected because it match the new id (${view})` });
		fmt.Printf("no match for future leader\n")
		return
	}

	if !term.validateViewChangeVotes(header.BlockHeight(), header.View(), viewChangeConfirmations) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", votes is invalid` });
		fmt.Printf("validateViewChangeVotes failed\n")
		return
	}

	if term.view > header.View() {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", view is from the past` });
		fmt.Printf("current view is higher than message view\n")
		return
	}

	if !ppMessageContent.SignedHeader().View().Equal(header.View()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", view doesn't match PP.view` });
		fmt.Printf("NewView.view and NewView.Preprepare.view do not match\n")
		return
	}

	if !ppMessageContent.SignedHeader().BlockHeight().Equal(header.BlockHeight()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", blockHeight doesn't match PP.blockHeight` });
		fmt.Printf("NewView.BlockHeight and NewView.Preprepare.BlockHeight do not match\n")
		return
	}

	latestVote := term.latestViewChangeVote(viewChangeConfirmations)
	if latestVote != nil {
		viewChangeMessageValid := term.isViewChangeValid(futureLeaderId, header.View(), latestVote)
		if !viewChangeMessageValid {
			//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", view change votes are invalid` });
			fmt.Printf("NewView.ViewChangeConfirmation (with latest view) is invalid\n")
			return
		}

		// rewrite this mess
		latestVoteBlockHash := latestVote.SignedHeader().PreparedProof().PreprepareBlockRef().BlockHash()
		if latestVoteBlockHash != nil {
			ppBlockHash := term.BlockUtils.CalculateBlockHash(nvm.Block())
			if !latestVoteBlockHash.Equal(ppBlockHash) {
				//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", the given _Block (PP._Block) doesn't match the best _Block from the VCProof` });
				fmt.Printf("NewView.ViewChangeConfirmation (with latest view) is invalid\n")
				return
			}
		}
	}

	ppm := &PreprepareMessage{
		content: ppMessageContent,
		block:   nvm.Block(),
	}

	if err := term.validatePreprepare(ppm); err == nil {
		term.newViewLocally = header.View()
		term.SetView(header.View())
		term.processPreprepare(ctx, ppm)
	}
}

func (term *LeanHelixTerm) QuorumSize() int {
	committeeMembersCount := len(term.committeeMembersPublicKeys)
	f := int(math.Floor(float64(committeeMembersCount-1) / 3))
	return committeeMembersCount - f
}

func (term *LeanHelixTerm) IsLeader() bool {
	return term.myPublicKey.Equal(term.leaderPublicKey)
}

func (term *LeanHelixTerm) countPrepared(height BlockHeight, view View, blockHash Uint256) int {
	return len(term.Storage.GetPrepareSendersPKs(height, view, blockHash))
}

func (term *LeanHelixTerm) isPreprepared(blockHeight BlockHeight, view View, blockHash Uint256) bool {
	ppm, ok := term.Storage.GetPreprepareMessage(blockHeight, view)
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

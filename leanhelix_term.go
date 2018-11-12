package leanhelix

import (
	"context"
	"fmt"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"math"
	"sort"
)

type leanHelixTerm struct {
	ctx context.Context
	KeyManager
	NetworkCommunication
	Storage
	electionTrigger ElectionTrigger
	BlockUtils
	messageFactory                *MessageFactory
	filter                        *ConsensusMessageFilter
	myPublicKey                   Ed25519PublicKey
	committeeMembersPublicKeys    []Ed25519PublicKey
	nonCommitteeMembersPublicKeys []Ed25519PublicKey
	height                        BlockHeight
	view                          View
	preparedLocally               bool
	committedBlock                Block
	leaderPublicKey               Ed25519PublicKey
	newViewLocally                View
}

func NewLeanHelixTerm(config *Config, filter *ConsensusMessageFilter, newBlockHeight BlockHeight) *leanHelixTerm {
	keyManager := config.KeyManager
	blockUtils := config.BlockUtils
	myPK := keyManager.MyPublicKey()
	comm := config.NetworkCommunication
	messageFactory := NewMessageFactory(keyManager)
	committeeMembers := comm.RequestOrderedCommittee(uint64(newBlockHeight))
	nonCommitteeMembers := make([]Ed25519PublicKey, 0)
	for _, member := range committeeMembers {
		if !member.Equal(myPK) {
			nonCommitteeMembers = append(nonCommitteeMembers, member)
		}
	}

	newTerm := &leanHelixTerm{
		height:                        newBlockHeight,
		KeyManager:                    keyManager,
		NetworkCommunication:          comm,
		Storage:                       config.Storage,
		electionTrigger:               config.ElectionTrigger,
		BlockUtils:                    blockUtils,
		committeeMembersPublicKeys:    committeeMembers,
		nonCommitteeMembersPublicKeys: nonCommitteeMembers,
		messageFactory:                messageFactory,
		myPublicKey:                   myPK,
		filter:                        filter,
	}

	return newTerm
}

func (term *leanHelixTerm) WaitForBlock(ctx context.Context) Block {
	term.startTerm(ctx)
	for {
		ctxWithElectionTrigger := term.electionTrigger.CreateElectionContext(ctx, term.view)
		message, err := term.filter.WaitForMessage(ctxWithElectionTrigger, term.height)

		if err != nil {
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

func (term *leanHelixTerm) startTerm(ctx context.Context) {
	term.initView(0)
	if term.IsLeader() {
		block := term.BlockUtils.RequestNewBlock(ctx, term.height)
		ppm := term.messageFactory.CreatePreprepareMessage(term.height, term.view, block)

		term.Storage.StorePreprepare(ppm)
		term.sendPreprepare(ctx, ppm)
	}
}

func (term *leanHelixTerm) handleMessage(ctx context.Context) {

}

func (term *leanHelixTerm) onReceivePreprepare(ctx context.Context, ppm *PreprepareMessage) {
	fmt.Println("onReceivePreprepare:", term.myPublicKey.KeyForMap(), "term", term.height)
	if term.validatePreprepare(ppm) {
		term.processPreprepare(ctx, ppm)
	}
}

func (term *leanHelixTerm) onReceivePrepare(ctx context.Context, pm *PrepareMessage) {
	fmt.Println("onReceivePrepare:", term.myPublicKey.KeyForMap(), "term", term.height)

	header := pm.content.SignedHeader()
	sender := pm.content.Sender()

	if !term.KeyManager.Verify(header.Raw(), sender) {
		fmt.Printf("verification failed for Prepare blockHeight=%v view=%v blockHash=%v", header.BlockHeight(), header.View(), header.BlockHash())
	}
	if term.view > header.View() {
		fmt.Printf("prepare view %v is less than OneHeight's view %v", header.View(), term.view)
	}
	if term.leaderPublicKey.Equal(sender.SenderPublicKey()) {
		fmt.Printf("prepare received from leader (only preprepare can be received from leader)")
	}
	term.Storage.StorePrepare(pm)
	if term.view == header.View() {
		term.checkPrepared(ctx, header.BlockHeight(), header.View(), header.BlockHash())
	}
}

func (term *leanHelixTerm) onReceiveCommit(cm *CommitMessage) {
	fmt.Println("onReceiveCommit:", term.myPublicKey.KeyForMap(), "term", term.height)
	header := cm.content.SignedHeader()
	sender := cm.content.Sender()

	if !term.KeyManager.Verify(header.Raw(), sender) {
		fmt.Printf("verification failed for Commit blockHeight=%v view=%v blockHash=%v", header.BlockHeight(), header.View(), header.BlockHash())
	}
	if term.view > header.View() {
		fmt.Printf("message Commit view %v is less than OneHeight's view %v", header.View(), term.view)
	}
	if term.leaderPublicKey.Equal(sender.SenderPublicKey()) {
		fmt.Printf("message Commit received from leader (only preprepare can be received from leader)")
	}
	term.Storage.StoreCommit(cm)
	if term.view == header.View() {
		term.checkCommitted(header.BlockHeight(), header.View(), header.BlockHash())
	}
}

func (term *leanHelixTerm) onReceiveViewChange(ctx context.Context, vcm *ViewChangeMessage) {
	fmt.Println("onReceiveViewChange:", term.myPublicKey.KeyForMap(), "term", term.height)

	header := vcm.content.SignedHeader()
	if !term.isViewChangeValid(term.myPublicKey, term.view, vcm.content) {
		fmt.Printf("message ViewChange is not valid")
	}
	if vcm.block == nil || header.PreparedProof() == nil {
		fmt.Printf("message ViewChange - block or prepared proof are nil")
	}
	calculatedBlockHash := term.BlockUtils.CalculateBlockHash(vcm.block)
	isValidDigest := calculatedBlockHash.Equal(header.PreparedProof().PreprepareBlockRef().BlockHash())
	if !isValidDigest {
		fmt.Printf("different block hashes for block provided with message, and the block provided by the PPM in the PreparedProof of the message")
	}
	term.Storage.StoreViewChange(vcm)
	term.checkElected(ctx, header.BlockHeight(), header.View())
}

func (term *leanHelixTerm) onReceiveNewView(ctx context.Context, nvm *NewViewMessage) {
	fmt.Println("onReceiveNewView:", term.myPublicKey.KeyForMap(), "term", term.height)

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
		fmt.Printf("verify failed")
	}

	futureLeaderId := term.calcLeaderPublicKey(header.View())
	if !sender.SenderPublicKey().Equal(futureLeaderId) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", rejected because it match the new id (${view})` });
		fmt.Printf("no match for future leader")
	}

	if !term.validateViewChangeConfirmations(header.BlockHeight(), header.View(), viewChangeConfirmations) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", votes is invalid` });
		fmt.Printf("validateViewChangeConfirmations failed")
	}

	if term.view > header.View() {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", view is from the past` });
		fmt.Printf("current view is higher than message view")
	}

	if !ppMessageContent.SignedHeader().View().Equal(header.View()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", view doesn't match PP.view` });
		fmt.Printf("NewView.view and NewView.Preprepare.view do not match")
	}

	if !ppMessageContent.SignedHeader().BlockHeight().Equal(header.BlockHeight()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", blockHeight doesn't match PP.blockHeight` });
		fmt.Printf("NewView.BlockHeight and NewView.Preprepare.BlockHeight do not match")
	}

	latestConfirmation := term.latestViewChangeConfirmation(viewChangeConfirmations)
	if latestConfirmation != nil {
		viewChangeMessageValid := term.isViewChangeValid(futureLeaderId, header.View(), latestConfirmation)
		if !viewChangeMessageValid {
			//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", view change votes are invalid` });
			fmt.Printf("NewView.ViewChangeConfirmation (with latest view) is invalid")
		}

		// rewrite this mess
		latestConfirmationPreprepareBlockHash := latestConfirmation.SignedHeader().PreparedProof().PreprepareBlockRef().BlockHash()
		if latestConfirmationPreprepareBlockHash != nil {
			ppBlockHash := term.BlockUtils.CalculateBlockHash(nvm.Block())
			if !latestConfirmationPreprepareBlockHash.Equal(ppBlockHash) {
				//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", the given _Block (PP._Block) doesn't match the best _Block from the VCProof` });
				fmt.Printf("NewView.ViewChangeConfirmation (with latest view) is invalid")
			}
		}
	}

	ppm := &PreprepareMessage{
		content: ppMessageContent,
		block:   nvm.Block(),
	}

	if term.validatePreprepare(ppm) {
		term.newViewLocally = header.View()
		term.SetView(header.View())
		term.processPreprepare(ctx, ppm)
	}
}

func (term *leanHelixTerm) validatePreprepare(ppm *PreprepareMessage) bool {

	blockHeight := ppm.BlockHeight()
	view := ppm.View()
	if term.hasPreprepare(blockHeight, view) {
		return false
	}

	header := ppm.Content().SignedHeader()
	sender := ppm.Content().Sender()
	if !term.KeyManager.Verify(header.Raw(), sender) {
		return false
	}

	leaderPublicKey := term.calcLeaderPublicKey(view)
	senderPublicKey := sender.SenderPublicKey()
	if !senderPublicKey.Equal(leaderPublicKey) {
		// Log
		return false
	}

	givenBlockHash := term.BlockUtils.CalculateBlockHash(ppm.Block())
	if !ppm.Content().SignedHeader().BlockHash().Equal(givenBlockHash) {
		return false
	}

	isValidBlock := term.BlockUtils.ValidateBlock(ppm.Block())

	if !isValidBlock {
		return false
	}

	return true
}

func (term *leanHelixTerm) hasPreprepare(blockHeight BlockHeight, view View) bool {
	_, ok := term.GetPreprepareMessage(blockHeight, view)
	return ok
}

func (term *leanHelixTerm) processPreprepare(ctx context.Context, ppm *PreprepareMessage) {
	header := ppm.content.SignedHeader()
	if term.view != header.View() {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], processPrePrepare, view doesn't match` });
		return
	}

	pm := term.messageFactory.CreatePrepareMessage(header.BlockHeight(), header.View(), header.BlockHash())
	term.Storage.StorePreprepare(ppm)
	term.Storage.StorePrepare(pm)
	term.sendPrepare(ctx, pm)
	term.checkPrepared(ctx, header.BlockHeight(), header.View(), header.BlockHash())
}

func (term *leanHelixTerm) GetView() View {
	return term.view
}

func (term *leanHelixTerm) sendPreprepare(ctx context.Context, message *PreprepareMessage) {
	rawMessage := message.ToConsensusRawMessage()
	term.NetworkCommunication.SendMessage(ctx, term.nonCommitteeMembersPublicKeys, rawMessage)
}

func (term *leanHelixTerm) sendPrepare(ctx context.Context, message *PrepareMessage) {
	rawMessage := message.ToConsensusRawMessage()
	term.NetworkCommunication.SendMessage(ctx, term.nonCommitteeMembersPublicKeys, rawMessage)
}

func (term *leanHelixTerm) sendCommit(ctx context.Context, message *CommitMessage) {
	rawMessage := message.ToConsensusRawMessage()
	term.NetworkCommunication.SendMessage(ctx, term.nonCommitteeMembersPublicKeys, rawMessage)
}

func (term *leanHelixTerm) initView(view View) {
	term.preparedLocally = false
	term.view = view
	term.leaderPublicKey = term.calcLeaderPublicKey(view)
}

func (term *leanHelixTerm) calcLeaderPublicKey(view View) Ed25519PublicKey {
	index := int(view) % len(term.committeeMembersPublicKeys)
	return term.committeeMembersPublicKeys[index]
}

func (term *leanHelixTerm) IsLeader() bool {
	return term.myPublicKey.Equal(term.leaderPublicKey)
}

func (term *leanHelixTerm) moveToNextLeader(ctx context.Context) {
	term.SetView(term.view + 1)
	preparedMessages := ExtractPreparedMessages(term.height, term.Storage, term.QuorumSize())
	vcm := term.messageFactory.CreateViewChangeMessage(term.height, term.view, preparedMessages)
	term.Storage.StoreViewChange(vcm)
	if term.IsLeader() {
		term.checkElected(ctx, term.height, term.view)
	} else {
		term.sendViewChange(ctx, vcm)
	}
}

func (term *leanHelixTerm) latestViewChangeConfirmation(confirmations []*ViewChangeMessageContent) *ViewChangeMessageContent {

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

func (term *leanHelixTerm) isViewChangeValid(targetLeaderPublicKey Ed25519PublicKey, view View, confirmation *ViewChangeMessageContent) bool {

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

	if !ValidatePreparedProof(term.height, newView, preparedProof, term.GetF(), term.KeyManager, term.committeeMembersPublicKeys, func(view View) Ed25519PublicKey { return term.calcLeaderPublicKey(view) }) {
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

func (term *leanHelixTerm) SetView(view View) {
	if term.view != view {
		term.initView(view)
	}
}

func (term *leanHelixTerm) GetF() int {
	return int(math.Floor(float64(len(term.committeeMembersPublicKeys))-1) / 3)
}

func (term *leanHelixTerm) validateViewChangeConfirmations(targetBlockHeight BlockHeight, targetView View, confirmations []*ViewChangeMessageContent) bool {

	minimumConfirmations := int(term.GetF()*2 + 1)

	if len(confirmations) < minimumConfirmations {
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

func (term *leanHelixTerm) QuorumSize() int {
	committeeMembersCount := len(term.committeeMembersPublicKeys)
	f := int(math.Floor(float64(committeeMembersCount-1) / 3))
	return committeeMembersCount - f
}

func (term *leanHelixTerm) checkElected(ctx context.Context, height BlockHeight, view View) {
	if term.newViewLocally < view {
		vcms := term.Storage.GetViewChangeMessages(height, view)
		minimumNodes := term.QuorumSize()
		if len(vcms) >= minimumNodes {
			term.onElected(ctx, view, vcms[:minimumNodes])
		}
	}
}

func (term *leanHelixTerm) onElected(ctx context.Context, view View, viewChangeMessages []*ViewChangeMessage) {
	term.newViewLocally = view
	term.SetView(view)
	block := GetLatestBlockFromViewChangeMessages(viewChangeMessages)
	if block == nil {
		block = term.BlockUtils.RequestNewBlock(term.ctx, term.height)
	}
	ppmContentBuilder := term.messageFactory.CreatePreprepareMessageContentBuilder(term.height, view, block)
	ppm := term.messageFactory.CreatePreprepareMessageFromContentBuilder(ppmContentBuilder, block)
	confirmations := extractConfirmationsFromViewChangeMessages(viewChangeMessages)
	nvm := term.messageFactory.CreateNewViewMessage(term.height, view, ppmContentBuilder, confirmations, block)
	term.Storage.StorePreprepare(ppm)
	term.sendNewView(ctx, nvm)
}

func (term *leanHelixTerm) sendNewView(ctx context.Context, nvm *NewViewMessage) {
	nvmRaw := nvm.ToConsensusRawMessage()
	term.NetworkCommunication.SendMessage(ctx, term.nonCommitteeMembersPublicKeys, nvmRaw)
	// log
}

func (term *leanHelixTerm) sendViewChange(ctx context.Context, viewChangeMessage *ViewChangeMessage) {

}

func (term *leanHelixTerm) checkPrepared(ctx context.Context, blockHeight BlockHeight, view View, blockHash Uint256) {
	if term.preparedLocally == false {
		if term.isPreprepared(blockHeight, view, blockHash) {
			countPrepared := term.countPrepared(blockHeight, view, blockHash)
			//const metaData = {
			//method: "checkPrepared",
			//	height: this.blockHeight,
			//		blockHash,
			//		countPrepared
			//};
			//this.logger.log({ subject: "Info", message: `counting`, metaData });
			if countPrepared >= term.QuorumSize()-1 {
				term.onPrepared(ctx, blockHeight, view, blockHash)
			}
		}
	}
}

func (term *leanHelixTerm) isPreprepared(blockHeight BlockHeight, view View, blockHash Uint256) bool {
	ppm, ok := term.Storage.GetPreprepareMessage(blockHeight, view)
	if !ok {
		return false
	}
	ppmBlock := ppm.block
	if ppmBlock == nil {
		return false
	}
	ppmBlockHash := ppmBlock.BlockHash() // TODO Use CalcBlockHash here (as in ts code)?
	//const metaData = {
	//method: "isPrePrepared",
	//	height: this.blockHeight,
	//		prePreparedBlockHash,
	//		blockHash,
	//		eq: prePreparedBlockHash.equals(blockHash)
	//};
	//this.logger.log({ subject: "Info", message: `isPrePrepared`, metaData });
	return ppmBlockHash.Equal(blockHash)
}

func (term *leanHelixTerm) countPrepared(height BlockHeight, view View, blockHash Uint256) int {
	return len(term.Storage.GetPrepareSendersPKs(height, view, blockHash))
}

func (term *leanHelixTerm) onPrepared(ctx context.Context, blockHeight BlockHeight, view View, blockHash Uint256) {
	term.preparedLocally = true
	cm := term.messageFactory.CreateCommitMessage(blockHeight, view, blockHash)
	term.Storage.StoreCommit(cm)
	term.sendCommit(ctx, cm)
	term.checkCommitted(blockHeight, view, blockHash)
}

func (term *leanHelixTerm) checkCommitted(blockHeight BlockHeight, view View, blockHash Uint256) {
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
		return
	}
	term.committedBlock = ppm.block
}

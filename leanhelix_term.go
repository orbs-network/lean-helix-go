package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"math"
	"sort"
	"strings"
)

type LeanHelixTerm interface {
	MessageReceiver
	GetView() View
}

type leanHelixTerm struct {
	ctx context.Context
	KeyManager
	NetworkCommunication
	Storage
	log             log.BasicLogger
	electionTrigger ElectionTrigger
	BlockUtils
	MyPublicKey                   Ed25519PublicKey
	CommitteeMembersPublicKeys    []Ed25519PublicKey
	NonCommitteeMembersPublicKeys []Ed25519PublicKey
	MessageFactory                *MessageFactory
	onCommittedBlock              func(block Block)
	height                        BlockHeight
	view                          View
	disposed                      bool
	preparedLocally               bool
	leaderPublicKey               Ed25519PublicKey
	newViewLocally                View
}

func NewLeanHelixTerm(ctx context.Context, config *TermConfig, newBlockHeight BlockHeight, onCommittedBlock func(block Block)) (LeanHelixTerm, error) {

	keyManager := config.KeyManager
	blockUtils := config.BlockUtils
	myPK := keyManager.MyPublicKey()
	comm := config.NetworkCommunication
	committeeMembers := comm.RequestOrderedCommittee(uint64(newBlockHeight))
	if len(committeeMembers) == 0 {
		return nil, fmt.Errorf("no members for _Block height %v", newBlockHeight)
	}
	nonCommitteeMembers := make([]Ed25519PublicKey, 0)
	for _, member := range committeeMembers {
		if !member.Equal(myPK) {
			nonCommitteeMembers = append(nonCommitteeMembers, member)
		}
	}

	newTerm := &leanHelixTerm{
		ctx:                        ctx,
		height:                     newBlockHeight,
		KeyManager:                 keyManager,
		NetworkCommunication:       comm,
		Storage:                    config.Storage,
		log:                        config.Logger.For(log.Service("leanhelix-height")),
		electionTrigger:            config.ElectionTrigger,
		BlockUtils:                 blockUtils,
		CommitteeMembersPublicKeys: committeeMembers,
		MessageFactory:             config.MessageFactory,
		onCommittedBlock:           onCommittedBlock,
		MyPublicKey:                myPK,
	}

	newTerm.startTerm(ctx)

	return newTerm, nil
}

func (term *leanHelixTerm) startTerm(ctx context.Context) {
	term.log.Info("StartTerm() ID=%s height=%d started", log.Stringable("my-id", term.KeyManager.MyPublicKey()), log.Stringable("height", term.height))
	term.initView(0)

	if !term.IsLeader() {
		term.log.Debug("StartTerm() is not leader, returning.", log.Stringable("id", term.KeyManager.MyPublicKey()), log.Stringable("height", term.height))
		return
	}
	term.log.Info("StartTerm() is leader", log.Stringable("id", term.KeyManager.MyPublicKey()), log.Stringable("height", term.height))
	// TODO This should _Block!!!
	block := term.BlockUtils.RequestNewBlock(ctx, term.height)
	term.log.Info("StartTerm() generated new _Block", log.Stringable("id", term.KeyManager.MyPublicKey()), log.Stringable("height", term.height), log.Stringable("_Block-hash", block.BlockHash()))
	if term.disposed {
		term.log.Debug("StartTerm() disposed, returning", log.Stringable("id", term.KeyManager.MyPublicKey()), log.Stringable("height", term.height))
		return
	}
	ppm := term.MessageFactory.CreatePreprepareMessage(term.height, term.view, block)

	term.Storage.StorePreprepare(ppm)
	term.sendPreprepare(ctx, ppm)

}

func (term *leanHelixTerm) OnReceivePreprepare(ctx context.Context, ppm *PreprepareMessage) error {
	ok := term.validatePreprepare(ppm)
	if !ok {
		panic("throw some error here") // TODO nicer error & log
	}
	term.processPreprepare(ppm)

	return nil
}

func (term *leanHelixTerm) OnReceivePrepare(ctx context.Context, pm *PrepareMessage) error {
	panic("not impl")
}

func (term *leanHelixTerm) OnReceiveCommit(ctx context.Context, cm *CommitMessage) error {
	panic("not impl")
}

func (term *leanHelixTerm) OnReceiveViewChange(ctx context.Context, vcm *ViewChangeMessage) error {
	panic("implement me")
}

func (term *leanHelixTerm) OnReceiveNewView(ctx context.Context, nvm *NewViewMessage) error {

	panic("convert ts->go")
	signedHeader := nvm.Content().SignedHeader()
	sender := nvm.Content().Sender()
	preprepareMessageContent := nvm.Content().PreprepareMessageContent()
	viewChangeConfirmationsIter := signedHeader.ViewChangeConfirmationsIterator()
	viewChangeConfirmations := make([]*ViewChangeMessageContent, 0, 1)
	for {
		if viewChangeConfirmationsIter.HasNext() {
			viewChangeConfirmations = append(viewChangeConfirmations, viewChangeConfirmationsIter.NextViewChangeConfirmations())
		}
	}

	if !term.KeyManager.Verify(signedHeader.Raw(), sender) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", ignored because the signature verification failed` });
		return fmt.Errorf("verify failed")
	}

	futureLeaderId := term.calcLeaderPublicKey(signedHeader.View())
	if !sender.SenderPublicKey().Equal(futureLeaderId) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", rejected because it match the new id (${view})` });
		return fmt.Errorf("no match for future leader")
	}

	if !term.validateViewChangeConfirmations(signedHeader.BlockHeight(), signedHeader.View(), viewChangeConfirmations) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", votes is invalid` });
		return fmt.Errorf("validateViewChangeConfirmations failed")
	}

	if term.view > signedHeader.View() {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", view is from the past` });
		return fmt.Errorf("current view is higher than message view")
	}

	if !preprepareMessageContent.SignedHeader().View().Equal(signedHeader.View()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", view doesn't match PP.view` });
		return fmt.Errorf("NewView.view and NewView.Preprepare.view do not match")
	}

	if !preprepareMessageContent.SignedHeader().BlockHeight().Equal(signedHeader.BlockHeight()) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", blockHeight doesn't match PP.blockHeight` });
		return fmt.Errorf("NewView.BlockHeight and NewView.Preprepare.BlockHeight do not match")
	}

	latestConfirmation := term.latestViewChangeConfirmation(viewChangeConfirmations)
	if latestConfirmation != nil {
		viewChangeMessageValid := term.isViewChangeValid(futureLeaderId, signedHeader.View(), latestConfirmation)
		if !viewChangeMessageValid {
			//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", view change votes are invalid` });
			return fmt.Errorf("NewView.ViewChangeConfirmation (with latest view) is invalid")
		}

		// rewrite this mess
		latestConfirmationPreprepareBlockHash := latestConfirmation.SignedHeader().PreparedProof().PreprepareBlockRef().BlockHash()
		if latestConfirmationPreprepareBlockHash != nil {
			ppBlockHash := term.BlockUtils.CalculateBlockHash(nvm.Block())
			if !latestConfirmationPreprepareBlockHash.Equal(ppBlockHash) {
				//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceiveNewView from "${senderPk}", the given _Block (PP._Block) doesn't match the best _Block from the VCProof` });
				return fmt.Errorf("NewView.ViewChangeConfirmation (with latest view) is invalid")
			}
		}
	}

	ppm := &PreprepareMessage{
		content: preprepareMessageContent,
		block:   nvm.Block(),
	}

	if term.validatePreprepare(ppm) {
		term.newViewLocally = signedHeader.View()
		term.SetView(signedHeader.View())
		term.processPreprepare(ppm)
	}

	return nil
}

func (term *leanHelixTerm) validatePreprepare(ppm *PreprepareMessage) bool {

	blockHeight := ppm.BlockHeight()
	view := ppm.View()
	if term.hasPreprepare(blockHeight, view) {
		term.log.Info("PPM already received", log.Stringable("_Block-height", blockHeight), log.Stringable("view", view))
		return false
	}
	if !term.KeyManager.Verify(ppm.Raw(), ppm.Content().Sender()) {
		term.log.Info("PPM did not pass verification") // TODO Elaborate
		return false
	}

	leaderPublicKey := term.calcLeaderPublicKey(view)

	if !ppm.Content().Sender().SenderPublicKey().Equal(leaderPublicKey) {
		// Log
		return false
	}

	givenBlockHash := term.BlockUtils.CalculateBlockHash(ppm.Block())
	if !ppm.Content().SignedHeader().BlockHash().Equal(givenBlockHash) {
		//term.log.Info({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceivePrePrepare from "${senderPk}", _Block rejected because it doesn't match the given blockHash (${view})` });
		return false
	}

	isValidBlock := term.BlockUtils.ValidateBlock(ppm.Block())
	if term.disposed {
		return false
	}

	if !isValidBlock {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${view}], onReceivePrePrepare from "${senderPk}", _Block is invalid` });
		return false
	}

	return true
}

func (term *leanHelixTerm) hasPreprepare(blockHeight BlockHeight, view View) bool {
	_, ok := term.GetPreprepareMessage(blockHeight, view)
	return ok
}

func (term *leanHelixTerm) processPreprepare(ppm *PreprepareMessage) {
	panic("impl me - create Prepare etc.")
}

func (term *leanHelixTerm) GetView() View {
	return term.view
}
func (term *leanHelixTerm) sendPreprepare(ctx context.Context, message *PreprepareMessage) {

	rawMessage := message.ToConsensusRawMessage()

	term.NetworkCommunication.SendMessage(ctx, term.NonCommitteeMembersPublicKeys, rawMessage)

	term.log.Debug("GossipSend preprepare",
		log.Stringable("senderPK", term.KeyManager.MyPublicKey()),
		log.String("targetPKs", pksToString(term.NonCommitteeMembersPublicKeys)),
		log.Stringable("height", message.View()),
		log.Stringable("blockHash", message.Content().SignedHeader().BlockHash()),
	)
}
func pksToString(keys []Ed25519PublicKey) string {
	pkStrings := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		pkStrings[i] = string(keys[i])
	}
	return strings.Join(pkStrings, ",")
}

func (term *leanHelixTerm) initView(view View) {
	term.preparedLocally = false
	term.view = view
	term.leaderPublicKey = term.calcLeaderPublicKey(view)
	term.electionTrigger.RegisterOnTrigger(view, func(v View) { term.onLeaderChange(v) })
}
func (term *leanHelixTerm) calcLeaderPublicKey(view View) Ed25519PublicKey {
	index := int(view) % len(term.CommitteeMembersPublicKeys)
	return term.CommitteeMembersPublicKeys[index]
}
func (term *leanHelixTerm) IsLeader() bool {
	return term.MyPublicKey.Equal(term.leaderPublicKey)
}
func (term *leanHelixTerm) onLeaderChange(counter View) {
	panic("not impl")
}

// TODO Unit-test this!!
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

	signedHeader := confirmation.SignedHeader()
	newView := signedHeader.View()
	preparedProof := signedHeader.PreparedProof()
	sender := confirmation.Sender()

	if !term.KeyManager.Verify(signedHeader.Raw(), sender) {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${newView}], onReceiveViewChange from "${senderPk}", ignored because the signature verification failed` });
		return false
	}

	if view > newView {
		//this.logger.log({ subject: "Warning", message: `blockHeight:[${blockHeight}], view:[${newView}], onReceiveViewChange from "${senderPk}", ignored because of unrelated view` });
		return false
	}

	if !ValidatePreparedProof(term.height, newView, preparedProof, term.GetF(), term.KeyManager, term.CommitteeMembersPublicKeys, func(view View) Ed25519PublicKey { return term.calcLeaderPublicKey(view) }) {
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
	return int(math.Floor(float64(len(term.CommitteeMembersPublicKeys))-1) / 3)
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

package leanhelix

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
	"strings"
)

type LeanHelixTerm interface {
	GetView() View
	OnReceivePreprepare(ppm PreprepareMessage)
}

type leanHelixTerm struct {
	KeyManager
	NetworkCommunication
	Storage
	log             log.BasicLogger
	electionTrigger ElectionTrigger
	BlockUtils
	MyPublicKey                   PublicKey
	CommitteeMembersPublicKeys    []PublicKey
	NonCommitteeMembersPublicKeys []PublicKey
	MessageFactory                MessageFactory
	onCommittedBlock              func(block Block)
	height                        BlockHeight
	view                          View
	disposed                      bool
	preparedLocally               bool
	leaderPublicKey               PublicKey
}

func NewLeanHelixTerm(config *TermConfig, newBlockHeight BlockHeight, onCommittedBlock func(block Block)) (LeanHelixTerm, error) {

	keyManager := config.KeyManager
	blockUtils := config.BlockUtils
	myPK := keyManager.MyPublicKey()
	comm := config.NetworkCommunication
	committeeMembers := comm.RequestOrderedCommittee(uint64(newBlockHeight))
	if len(committeeMembers) == 0 {
		return nil, fmt.Errorf("no members for block height %v", newBlockHeight)
	}
	nonCommitteeMembers := make([]PublicKey, 0)
	for _, member := range committeeMembers {
		if !member.Equals(myPK) {
			nonCommitteeMembers = append(nonCommitteeMembers, member)
		}
	}

	newTerm := &leanHelixTerm{
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

	newTerm.startTerm()

	return newTerm, nil
}

func (term *leanHelixTerm) startTerm() {
	term.log.Info("StartTerm() ID=%s height=%d started", log.Stringable("my-id", term.KeyManager.MyPublicKey()), log.Stringable("height", term.height))
	term.initView(0)

	if !term.IsLeader() {
		term.log.Debug("StartTerm() is not leader, returning.", log.Stringable("id", term.KeyManager.MyPublicKey()), log.Stringable("height", term.height))
		return
	}
	term.log.Info("StartTerm() is leader", log.Stringable("id", term.KeyManager.MyPublicKey()), log.Stringable("height", term.height))
	// TODO This should block!!!
	block := term.BlockUtils.RequestNewBlock(term.height)
	term.log.Info("StartTerm() generated new block", log.Stringable("id", term.KeyManager.MyPublicKey()), log.Stringable("height", term.height), log.Stringable("block-hash", block.GetBlockHash()))
	if term.disposed {
		term.log.Debug("StartTerm() disposed, returning", log.Stringable("id", term.KeyManager.MyPublicKey()), log.Stringable("height", term.height))
		return
	}
	ppm := term.MessageFactory.CreatePreprepareMessage(term.height, term.view, block)
	term.Storage.StorePreprepare(ppm)
	term.sendPreprepare(ppm)

}

func (term *leanHelixTerm) OnReceivePreprepare(ppm PreprepareMessage) {
	panic("implement me")
}

func (term *leanHelixTerm) GetView() View {
	return term.view
}
func (term *leanHelixTerm) sendPreprepare(message PreprepareMessage) {
	term.NetworkCommunication.SendPreprepare(term.NonCommitteeMembersPublicKeys, message)

	term.log.Debug("GossipSend preprepare",
		log.Stringable("senderPK", term.KeyManager.MyPublicKey()),
		log.String("targetPKs", pksToString(term.NonCommitteeMembersPublicKeys)),
		log.Stringable("height", message.SignedHeader().View()),
		log.Stringable("blockHash", message.SignedHeader().BlockHash()),
	)
}
func pksToString(keys []PublicKey) string {
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
func (term *leanHelixTerm) calcLeaderPublicKey(view View) PublicKey {
	index := int(view) % len(term.CommitteeMembersPublicKeys)
	return term.CommitteeMembersPublicKeys[index]
}
func (term *leanHelixTerm) IsLeader() bool {
	return term.MyPublicKey.Equals(term.leaderPublicKey)
}
func (term *leanHelixTerm) onLeaderChange(counter View) {
	panic("not impl")
}

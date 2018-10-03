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
	MyPublicKey            PublicKey
	TermMembersPublicKeys  []PublicKey
	OtherMembersPublicKeys []PublicKey
	MessageFactory         MessageFactory
	onCommittedBlock       func(block Block)
	height                 BlockHeight
	view                   View
	disposed               bool
	preparedLocally        bool
	leaderPublicKey        PublicKey
}

func NewLeanHelixTerm(config *TermConfig, newHeight BlockHeight, onCommittedBlock func(block Block)) (LeanHelixTerm, error) {

	keyManager := config.KeyManager
	blockUtils := config.BlockUtils
	myPK := keyManager.MyPublicKey()
	comm := config.NetworkCommunication
	termMembers := comm.GetMembersPKs(uint64(newHeight))
	if len(termMembers) == 0 {
		return nil, fmt.Errorf("no members for block height %v", newHeight)
	}
	otherMembers := make([]PublicKey, 0)
	for _, member := range termMembers {
		if !member.Equals(myPK) {
			otherMembers = append(otherMembers, member)
		}
	}

	newTerm := &leanHelixTerm{
		height:                newHeight,
		KeyManager:            keyManager,
		NetworkCommunication:  comm,
		Storage:               config.Storage,
		log:                   config.Logger.For(log.Service("leanhelix-term")),
		electionTrigger:       config.ElectionTrigger,
		BlockUtils:            blockUtils,
		TermMembersPublicKeys: termMembers,
		MessageFactory:        NewMessageFactory(blockUtils.CalculateBlockHash, keyManager),
		onCommittedBlock:      onCommittedBlock,
		MyPublicKey:           myPK,
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
	term.NetworkCommunication.SendPreprepare(term.OtherMembersPublicKeys, message)

	term.log.Debug("GossipSend preprepare",
		log.Stringable("senderPK", term.KeyManager.MyPublicKey()),
		log.String("targetPKs", pksToString(term.OtherMembersPublicKeys)),
		log.Stringable("height", message.View()),
		log.Stringable("blockHash", message.BlockHash()),
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
	index := int(view) % len(term.TermMembersPublicKeys)
	return term.TermMembersPublicKeys[index]
}
func (term *leanHelixTerm) IsLeader() bool {
	return term.MyPublicKey.Equals(term.leaderPublicKey)
}
func (term *leanHelixTerm) onLeaderChange(counter View) {
	panic("not impl")
}

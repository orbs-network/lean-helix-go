package leanhelix

type LeanHelixTerm interface {
	GetView() ViewCounter
	OnReceivePreprepare(ppm PreprepareMessage)
}

type leanHelixTerm struct {
	KeyManager
	NetworkCommunication
	Storage
	Logger
	ElectionTrigger
	BlockUtils
	MyPublicKey            PublicKey
	TermMembersPublicKeys  []PublicKey
	OtherMembersPublicKeys []PublicKey
	MessageFactory         MessageFactory
	onCommittedBlock       func(block Block)
	height                 BlockHeight
	view                   ViewCounter
}

func NewLeanHelixTerm(config *TermConfig, newHeight BlockHeight, onCommittedBlock func(block Block)) LeanHelixTerm {

	keyManager := config.KeyManager
	blockUtils := config.BlockUtils
	comm := config.NetworkCommunication
	termMembers := comm.GetMembersPKs(uint64(newHeight))
	otherMembers := make([]PublicKey, 0)
	for _, member := range termMembers {
		if !member.Equals(keyManager.MyID()) {
			otherMembers = append(otherMembers, member)
		}
	}

	newTerm := &leanHelixTerm{
		height:                newHeight,
		KeyManager:            keyManager,
		NetworkCommunication:  comm,
		Storage:               config.Storage,
		Logger:                config.Logger,
		ElectionTrigger:       config.ElectionTrigger,
		BlockUtils:            blockUtils,
		TermMembersPublicKeys: termMembers,
		MessageFactory:        NewMessageFactory(blockUtils.CalculateBlockHash, keyManager),
		onCommittedBlock:      onCommittedBlock,
	}

	newTerm.startTerm()

	return newTerm
}

func (term *leanHelixTerm) startTerm() {
	term.Logger.Info("StartTerm() ID=%s height=%d started", term.KeyManager.MyID(), term.height)
	term.initView(0)

	if !term.IsLeader() {
		term.Logger.Debug("On StartTerm ID=%s height=%d - not leader, returning", term.KeyManager.MyID(), term.height)
		return
	}
	term.Logger.Info("StartTerm() ID=%s height=%d is leader", term.KeyManager.MyID(), term.height)
	// TODO This should block!!!
	block := term.BlockUtils.RequestNewBlock(term.height)
	term.Logger.Info("StartTerm() ID=%s height=%d generated new block with hash=%s", term.KeyManager.MyID(), term.height, block.Header().BlockHash())
	if term.disposed {
		term.Logger.Debug("On StartTerm ID=%s height=%d - disposed, returning", term.KeyManager.MyID(), term.height)
		return
	}
	ppm := term.MessageFactory.CreatePreprepareMessage(term.height, term.view, block)
	term.Storage.StorePreprepare(ppm)
	term.sendPreprepare(ppm)

}

func (term *leanHelixTerm) OnReceivePreprepare(ppm PreprepareMessage) {
	panic("implement me")
}

func (term *leanHelixTerm) GetView() ViewCounter {
	return term.view
}
func (term *leanHelixTerm) sendPreprepare(message PreprepareMessage) {
	term.NetworkCommunication.SendPreprepare(term.OtherMembersPublicKeys, message)

	data := make(LogData)
	data["senderPK"] = string(term.KeyManager.MyID())
	data["targetPKs"] = string(term.OtherMembersPublicKeys)
	data["height"] = message.Term()
	data["view"] = message.View()
	data["blockHash"] = message.BlockHash()

	term.Logger.Debug("GossipSend preprepare", data)
}

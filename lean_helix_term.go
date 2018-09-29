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
	view                   ViewCounter
}

func NewLeanHelixTerm(config *TermConfig, height BlockHeight, onCommittedBlock func(block Block)) LeanHelixTerm {

	keyManager := config.KeyManager
	blockUtils := config.BlockUtils
	comm := config.NetworkCommunication
	termMembers := comm.GetMembersPKs(uint64(height))
	otherMembers := make([]PublicKey, 0)
	for _, member := range termMembers {
		if !member.Equals(keyManager.MyID()) {
			otherMembers = append(otherMembers, member)
		}
	}

	newTerm := &leanHelixTerm{
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
	panic("not impl")
}

func (term *leanHelixTerm) OnReceivePreprepare(ppm PreprepareMessage) {
	panic("implement me")
}

func (term *leanHelixTerm) GetView() ViewCounter {
	return term.view
}

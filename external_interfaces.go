package leanhelix

// TODO Is this only for testing, or is there real need for it in code?
type Block interface {
	GetHeight() BlockHeight
	GetBlockHash() BlockHash
	//Body() []byte
}

type NetworkCommunication interface {
	SendToMembers(publicKeys []PublicKey, messageType string, message []MessageTransporter)

	// Copied from TS code as is
	RequestOrderedCommittee(seed uint64) []PublicKey
	IsMember(pk PublicKey) bool

	SendPreprepare(publicKeys []PublicKey, message PreprepareMessage)
	SendPrepare(publicKeys []PublicKey, message PrepareMessage)
	SendCommit(publicKeys []PublicKey, message CommitMessage)
	SendViewChange(publicKey PublicKey, message ViewChangeMessage) // TODO Is this ok to be single pk? (see NetworkCommunication.ts)
	SendNewView(publicKeys []PublicKey, message NewViewMessage)

	RegisterToPreprepare(cb func(message PreprepareMessage))
	RegisterToPrepare(cb func(message PrepareMessage))
	RegisterToCommit(cb func(message CommitMessage))
	RegisterToViewChange(cb func(message ViewChangeMessage))
	RegisterToNewView(cb func(message NewViewMessage))
}

// TODO Maybe KeyManager shouldn't hold MyPublicKey and just be a static singleton that accepts ID like every other property
type KeyManager interface {
	SignBlockRef(blockRef BlockRef) SenderSignature // TODO uses its internal ID to sign, is it ok? probably yes
	SignViewChange(vcHeader ViewChangeHeader) SenderSignature
	SignNewView(nvHeader NewViewHeader) SenderSignature

	VerifyBlockRef(blockRef BlockRef, sender SenderSignature) bool // TODO this accepts SignatureRef.Sender() - this is smelly because SignBlockRef() doesn't explicitly accept PK snd VerifyBlockRef() does.
	VerifyViewChange(vcHeader ViewChangeHeader, sender SenderSignature) bool
	VerifyNewView(nvHeader NewViewHeader, sender SenderSignature) bool

	MyPublicKey() PublicKey
}

// TODO Maybe BlockHandler is better name? or BlockService
type BlockUtils interface {
	// Does Commit() go here?
	CalculateBlockHash(block Block) BlockHash
	RequestNewBlock(blockHeight BlockHeight) Block
	ValidateBlock()
	RequestCommittee()
}

type MessageFactory interface {
	CreatePreprepareMessage(blockHeight BlockHeight, view View, block Block) PreprepareMessage
	CreatePrepareMessage(blockHeight BlockHeight, view View, blockHash BlockHash) PrepareMessage
	CreateCommitMessage(blockHeight BlockHeight, view View, blockHash BlockHash) CommitMessage
	CreateViewChangeMessage(blockHeight BlockHeight, view View, preparedMessages []PreprepareMessage) ViewChangeMessage
	CreateNewViewMessage(blockHeight BlockHeight, view View, preprepareMessage PreprepareMessage, viewChangeConfirmations []ViewChangeConfirmation) NewViewMessage
	//CreatePreparedProof(preprepare PreprepareMessage, prepares []PrepareMessage) PreparedProof
}

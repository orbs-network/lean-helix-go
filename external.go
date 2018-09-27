package leanhelix

// External interfaces of this library (temporary)

type BlockHeight uint64
type ViewCounter uint64
type BlockHash []byte

func (hash BlockHash) Equals(other BlockHash) bool {
	return string(hash) == string(other)
}

type PublicKey []byte

func (pk PublicKey) Equals(other PublicKey) bool {
	return string(pk) == string(other)
}

type Signature []byte

func (s Signature) Equals(other Signature) bool {
	return string(s) == string(other)
}

type MessageType string

const (
	MESSAGE_TYPE_PREPREPARE  MessageType = "preprepare"
	MESSAGE_TYPE_PREPARE     MessageType = "prepare"
	MESSAGE_TYPE_COMMIT      MessageType = "commit"
	MESSAGE_TYPE_VIEW_CHANGE MessageType = "view-change"
	MESSAGE_TYPE_NEW_VIEW    MessageType = "new-view"
)

type BlockHeader interface {
	Term() BlockHeight
	BlockHash() BlockHash
}

// TODO Is this only for testing, or is there real need for it in code?
type Block interface {
	Header() BlockHeader
	Body() []byte
}

type NetworkCommunication interface {
	SendToMembers(publicKeys []PublicKey, messageType string, message []MessageTransporter)

	// Copied from TS code as is
	GetMembersPKs(seed int) []string
	IsMember(pk PublicKey) bool

	SendPreprepare(pks []PublicKey, message PreprepareMessage)
	SendPrepare(pks []PublicKey, message PrepareMessage)
	SendCommit(pks []PublicKey, message CommitMessage)
	SendViewChange(pk PublicKey, message ViewChangeMessage) // TODO Is this ok to be single pk? (see NetworkCommunication.ts)
	SendNewView(pks []PublicKey, message NewViewMessage)

	RegisterToPreprepare(cb func(message PreprepareMessage))
	RegisterToPrepare(cb func(message PrepareMessage))
	RegisterToCommit(cb func(message CommitMessage))
	RegisterToViewChange(cb func(message ViewChangeMessage))
	RegisterToNewView(cb func(message NewViewMessage))
}

// TODO Maybe KeyManager shouldn't hold MyID and just be a static singleton that accepts ID like every other property
type KeyManager interface {
	SignBlockRef(blockRef BlockRef) SenderSignature // TODO uses its internal ID to sign
	SignViewChange(vcm ViewChangeMessage) SenderSignature
	SignNewView(nvm NewViewMessage) SenderSignature

	VerifyBlockRef(blockRef BlockRef, sender SenderSignature) bool // TODO this accepts SignatureRef.Sender() - this is smelly because SignBlockRef() doesn't explicitly accept PK snd VerifyBlockRef() does.
	VerifyViewChange(vcm ViewChangeMessage, sender SenderSignature) bool
	VerifyNewView(nvm NewViewMessage, sender SenderSignature) bool

	MyID() PublicKey
}

// TODO Maybe BlockHandler is better name? or BlockService
type BlockUtils interface {
	CalculateBlockHash(block Block) BlockHash
	RequestNewBlock()
	ValidateBlock()
	RequestCommittee()
}

type SenderSignature interface {
	SenderPublicKey() PublicKey
	Signature() Signature
}

type HasMessageType interface {
	MessageType() MessageType
}

type MessageTransporter interface {
	SenderSignature
	HasMessageType
}

type BlockRef interface {
	HasMessageType
	Term() BlockHeight
	View() ViewCounter
	BlockHash() BlockHash
}

// TODO this is different from definition of LeanHelixPreparedProof in lean_helix.mb.go:448 in orbs-spec
type PreparedProof interface {
	PreprepareMessage() PreprepareMessage
	PrepareMessages() []PrepareMessage
}

// TODO refactor
type PreparedProofInternal struct {
	preprepare PreprepareMessage
	prepares   []PrepareMessage
}

func (pf *PreparedProofInternal) PreprepareMessage() PreprepareMessage {
	return pf.preprepare
}

func (pf *PreparedProofInternal) PrepareMessages() []PrepareMessage {
	return pf.prepares
}

type PreprepareMessage interface {
	BlockRef
	Sender() SenderSignature
	Block() Block
}

//////

type Adapter interface {
	RequestNewBlock()
	CommitBlock(Block)
}

////

type PrepareMessage interface {
	BlockRef
	Sender() SenderSignature
}

type CommitMessage interface {
	BlockRef
	Sender() SenderSignature
	// TODO Add RandomSeedShare?
}

type ViewChangeMessage interface {
	BlockRef
	Sender() SenderSignature
	Block() Block
	PreparedProof() PreparedProof
}

type NewViewMessage interface {
	BlockRef // TODO doesn't need BlockHash so maybe replace with Term() and View()
	ViewChangeConfirmations() []ViewChangeConfirmation
}

type ViewChangeConfirmation interface {
	Sender() SenderSignature
	ViewChangeMessage() ViewChangeMessage
}

func CreatePreparedProof(ppm PreprepareMessage, pms []PrepareMessage) PreparedProof {
	return &PreparedProofInternal{
		preprepare: ppm,
		prepares:   pms,
	}
}

package leanhelix

// External interfaces of this library (temporary)

type BlockHeight uint64

func (h BlockHeight) String() string {
	return string(h)
}

type View uint64

func (v View) String() string {
	return string(v)
}

type BlockHash []byte

func (hash BlockHash) String() string {
	return string(hash)
}

func (hash BlockHash) Equals(other BlockHash) bool {
	return string(hash) == string(other)
}

type PublicKey []byte

func (pk PublicKey) String() string {
	return string(pk)
}
func (pk PublicKey) Equals(other PublicKey) bool {
	return string(pk) == string(other)
}

type Signature []byte

func (s Signature) String() string {
	return string(s)
}
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
}

// TODO Is this only for testing, or is there real need for it in code?
type Block interface {
	GetTerm() BlockHeight
	GetBlockHash() BlockHash

	//Body() []byte
}

type NetworkCommunication interface {
	SendToMembers(publicKeys []PublicKey, messageType string, message []MessageTransporter)

	// Copied from TS code as is
	GetMembersPKs(seed uint64) []PublicKey
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
	SignBlockRef(blockRef BlockRef) SenderSignature // TODO uses its internal ID to sign
	SignViewChange(vcm ViewChangeMessage) SenderSignature
	SignNewView(nvm NewViewMessage) SenderSignature

	VerifyBlockRef(blockRef BlockRef, sender SenderSignature) bool // TODO this accepts SignatureRef.Sender() - this is smelly because SignBlockRef() doesn't explicitly accept PK snd VerifyBlockRef() does.
	VerifyViewChange(vcm ViewChangeMessage, sender SenderSignature) bool
	VerifyNewView(nvm NewViewMessage, sender SenderSignature) bool

	MyPublicKey() PublicKey
}

// TODO Maybe BlockHandler is better name? or BlockService
type BlockUtils interface {
	// Does Commit() go here?
	CalculateBlockHash(block Block) BlockHash
	RequestNewBlock(height BlockHeight) Block
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
	View() View
	BlockHash() BlockHash // TODO Gad: rename this to "current block hash"???
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

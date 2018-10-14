package leanhelix

import "github.com/orbs-network/orbs-network-go/services/consensusalgo/leanhelix"

type HasMessageType interface {
	MessageType() MessageType
}

type Serializable interface {
	Raw() []byte
}

type MessageTransporter interface {
	HasMessageType
}

type MessageContent interface {
	HasMessageType
	Serializable
	SignedHeader() BlockRef
	Sender() SenderSignature
}

// PP
type PreprepareMessage interface {
	MessageContent
	Block
}

type PrepareMessage interface {
	MessageContent
}

type CommitMessage interface {
	MessageContent
}

type ViewChangeMessage interface {
	MessageContent
	Block
}

type NewViewMessage interface {
	MessageContent
}

type preprepareMessage struct {
	Content *PreprepareMessageContent
	Block
}

func (ppm *preprepareMessage) SignedHeader() BlockRef {
	return ppm.SignedHeader()
}

func (ppm *preprepareMessage) Sender() SenderSignature {
	return ppm.Sender()
}

func (ppm *preprepareMessage) MessageType() MessageType {
	return LEAN_HELIX_PREPREPARE
}

func (ppm *preprepareMessage) Raw() []byte {
	return ppm.Raw()
}

type prepareMessage struct {
	Content *PrepareMessageContent
}

func (pm *prepareMessage) MessageType() MessageType {
	return LEAN_HELIX_PREPARE
}

func (pm *prepareMessage) Raw() []byte {
	return pm.Raw()
}

type commitMessage struct {
	Content *CommitMessageContent
}

func (cm *commitMessage) MessageType() MessageType {
	return LEAN_HELIX_COMMIT
}

func (cm *commitMessage) Raw() []byte {
	return cm.Raw()
}

type viewChangeMessage struct {
	Content *ViewChangeMessageContent
	Block
}

func (vcm *viewChangeMessage) MessageType() MessageType {
	return LEAN_HELIX_VIEW_CHANGE
}

func (vcm *viewChangeMessage) Raw() []byte {
	return vcm.Raw()
}

type newViewMessage struct {
	Content *NewViewMessageContent
}

func (nvm *newViewMessage) MessageType() MessageType {
	return LEAN_HELIX_NEW_VIEW
}

func (nvm *newViewMessage) Raw() []byte {
	return nvm.Raw()
}

type MessageFactory interface {
	// Message creation methods

	//CreatePreprepareMessage(blockRef BlockRef, sender SenderSignature, block Block) PreprepareMessage
	CreatePreprepareMessage(blockHeight uint64, view uint64, block Block) PreprepareMessage
	CreatePrepareMessage(blockRef BlockRef, sender SenderSignature) PrepareMessage
	CreateCommitMessage(blockRef BlockRef, sender SenderSignature) CommitMessage
	CreateViewChangeMessage(vcHeader ViewChangeHeader, sender SenderSignature, block Block) ViewChangeMessage
	CreateNewViewMessage(preprepareMessage PreprepareMessage, nvHeader NewViewHeader, sender SenderSignature) NewViewMessage

	// Auxiliary methods

	//CreateSenderSignature(sender []byte, signature []byte) SenderSignature
	//CreateBlockRef(messageType int, blockHeight int, view int, blockHash []byte) BlockRef
	//CreateNewViewHeader(messageType int, blockHeight int, view int, confirmations []ViewChangeConfirmation) NewViewHeader
	//CreateViewChangeConfirmation(vcHeader ViewChangeHeader, sender SenderSignature) ViewChangeConfirmation
	//CreateViewChangeHeader(blockHeight int, view int, proof PreparedProof) ViewChangeHeader
	//CreatePreparedProof(ppBlockRef BlockRef, pBlockRef BlockRef, ppSender SenderSignature, pSenders []SenderSignature) PreparedProof

	// TODO Remove old methods once not needed
}

type InternalMessageFactory interface {
	// TODO USe TDD to decide on methods here
}

//type MessageType string
//
//const (
//	MESSAGE_TYPE_PREPREPARE  MessageType = "preprepare"
//	MESSAGE_TYPE_PREPARE     MessageType = "prepare"
//	MESSAGE_TYPE_COMMIT      MessageType = "commit"
//	MESSAGE_TYPE_VIEW_CHANGE MessageType = "view-change"
//	MESSAGE_TYPE_NEW_VIEW    MessageType = "new-view"
//)

//type Message interface {
//	SignaturePair() SignaturePair
//}
//
//type BlockRefMessage struct {
//	SignaturePair *SignaturePair
//	Content       *BlockMessageContent
//}
//
//type PrePrepareMessage struct {
//	*BlockRefMessage
//	Block *primitives.Block
//}
//
//type PrepareMessage struct {
//	*BlockRefMessage
//}
//
//type CommitMessage struct {
//	*BlockRefMessage
//}
//
//type ViewChangeMessage struct {
//	Content       *ViewChangeMessageContent
//	SignaturePair *SignaturePair
//	Block         *primitives.Block // optional
//}
//
//type NewViewMessage struct {
//	SignaturePair     *SignaturePair
//	PreprepareMessage *PrePrepareMessage
//	Content           *NewViewContent
//}
//
//type NewViewContent struct {
//	MessageType   MessageType
//	BlockHeight          primitives.BlockHeight
//	View          primitives.View
//	Confirmations []*ViewChangeConfirmation
//}
//
//type ViewChangeConfirmation struct {
//	Content       *ViewChangeMessageContent
//	SignaturePair *SignaturePair
//}
//
//type PreparedMessages struct {
//	PreprepareMessage *PrePrepareMessage
//	PrepareMessages   []*PrepareMessage
//}
//
//type SignaturePair struct {
//	SignerPublicKey  primitives.PublicKey
//	ContentSignature string
//}
//
//type BlockMessageContent struct {
//	MessageType MessageType
//	BlockHeight        primitives.BlockHeight
//	View        primitives.View
//	BlockHash   primitives.BlockHash
//}
//
//type ViewChangeMessageContent struct {
//	MessageType   MessageType
//	BlockHeight          primitives.BlockHeight
//	View          primitives.View
//	PreparedProof *PreparedProof
//}
//
//type PreparedProof struct {
//	PreprepareBlockRefMessage *PrePrepareMessage
//	PrepareBlockRefMessages   []*PrepareMessage
//}

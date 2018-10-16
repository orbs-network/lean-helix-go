package leanhelix

import "github.com/orbs-network/lean-helix-go/primitives"

type HasMessageType interface {
	MessageType() MessageType
}

type Serializable interface {
	Raw() []byte
}

type MessageTransporter interface {
	HasMessageType
	Sender() *SenderSignature
}

type MessageContent interface {
	HasMessageType
	Serializable
	SignedHeader() *BlockRef
	Sender() *SenderSignature
}

// PP
type PreprepareMessage interface {
	MessageContent
	Block() Block
}

type PrepareMessage interface {
	MessageContent
}

type CommitMessage interface {
	MessageContent
}

type ViewChangeMessage interface {
	HasMessageType
	Serializable
	SignedHeader() *ViewChangeHeader
	Sender() *SenderSignature
	Block
}

type NewViewMessage interface {
	HasMessageType
	Serializable
	SignedHeader() *NewViewHeader
	Sender() *SenderSignature
	PreprepareMessage() PreprepareMessage
}

type PreprepareMessageImpl struct {
	Content *PreprepareMessageContent
	MyBlock Block
}

func (ppm *PreprepareMessageImpl) MessageType() MessageType {
	return LEAN_HELIX_PREPREPARE
}

func (ppm *PreprepareMessageImpl) SignedHeader() *BlockRef {
	return ppm.SignedHeader()
}

func (ppm *PreprepareMessageImpl) Sender() *SenderSignature {
	return ppm.Sender()
}

func (ppm *PreprepareMessageImpl) Raw() []byte {
	return ppm.Content.Raw()
}

func (ppm *PreprepareMessageImpl) Block() Block {
	return ppm.MyBlock
}

type PrepareMessageImpl struct {
	Content *PrepareMessageContent
}

func (pm *PrepareMessageImpl) MessageType() MessageType {
	return LEAN_HELIX_PREPARE
}

func (pm *PrepareMessageImpl) SignedHeader() *BlockRef {
	return pm.SignedHeader()
}

func (pm *PrepareMessageImpl) Sender() *SenderSignature {
	return pm.Sender()
}

func (pm *PrepareMessageImpl) Raw() []byte {
	return pm.Content.Raw()
}

type CommitMessageImpl struct {
	Content *CommitMessageContent
}

func (cm *CommitMessageImpl) MessageType() MessageType {
	return LEAN_HELIX_COMMIT
}

func (cm *CommitMessageImpl) SignedHeader() *BlockRef {
	return cm.SignedHeader()
}

func (cm *CommitMessageImpl) Sender() *SenderSignature {
	return cm.Sender()
}

func (cm *CommitMessageImpl) Raw() []byte {
	return cm.Content.Raw()
}

type ViewChangeMessageImpl struct {
	Content *ViewChangeMessageContent
	Block
}

func (vcm *ViewChangeMessageImpl) MessageType() MessageType {
	return LEAN_HELIX_VIEW_CHANGE
}

func (vcm *ViewChangeMessageImpl) SignedHeader() *ViewChangeHeader {
	return vcm.SignedHeader()
}

func (vcm *ViewChangeMessageImpl) Sender() *SenderSignature {
	return vcm.Sender()
}

func (vcm *ViewChangeMessageImpl) Raw() []byte {
	return vcm.Content.Raw()
}

type newViewMessage struct {
	Content *NewViewMessageContent
}

func (nvm *newViewMessage) SignedHeader() BlockRef {
	return nvm.SignedHeader()
}

func (nvm *newViewMessage) Sender() SenderSignature {
	return nvm.Sender()
}

func (nvm *newViewMessage) PreprepareMessage() PreprepareMessage {
	return nvm.PreprepareMessage()
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
	CreatePreprepareMessage(blockHeight primitives.BlockHeight, view primitives.View, block Block) PreprepareMessage
	CreatePrepareMessage(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.Uint256) PrepareMessage
	CreateCommitMessage(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.Uint256) CommitMessage
	// TODO Add PreparedMessages
	CreateViewChangeMessage(blockHeight primitives.BlockHeight, view primitives.View, preparedMessages *PreparedMessages) ViewChangeMessage
	CreateNewViewMessage(blockHeight primitives.BlockHeight, view primitives.View, ppm PreprepareMessage, confirmations []ViewChangeConfirmation) NewViewMessage

	// Auxiliary methods

	//CreateSenderSignature(sender []byte, signature []byte) SenderSignature
	//CreateBlockRef(messageType int, blockHeight int, view int, blockHash []byte) BlockRef
	//CreateNewViewHeader(messageType int, blockHeight int, view int, confirmations []ViewChangeConfirmation) NewViewHeader
	//CreateViewChangeConfirmation(vcHeader ViewChangeHeader, sender SenderSignature) ViewChangeConfirmation
	//CreateViewChangeHeader(blockHeight int, view int, proof PreparedProof) ViewChangeHeader
	//CreatePreparedProof(ppBlockRef BlockRef, pBlockRef BlockRef, ppSender SenderSignature, pSenders []SenderSignature) PreparedProof

	// TODO Remove old methods once not needed
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

package leanhelix

import (
	. "github.com/orbs-network/lean-helix-go/primitives"
)

type MessageFactory interface {
	// Message creation methods
	CreatePreprepareMessage(blockHeight BlockHeight, view View, block Block) PreprepareMessage
	CreatePrepareMessage(blockHeight BlockHeight, view View, blockHash Uint256) PrepareMessage
	CreateCommitMessage(blockHeight BlockHeight, view View, blockHash Uint256) CommitMessage
	CreateViewChangeMessage(blockHeight BlockHeight, view View, preparedMessages *PreparedMessages) ViewChangeMessage
	CreateNewViewMessage(blockHeight BlockHeight, view View, ppmcb *PreprepareMessageContentBuilder, confirmations []*ViewChangeMessageContentBuilder, block Block) NewViewMessage

	// Helper methods
	CreatePreprepareMessageContentBuilder(blockHeight BlockHeight, view View, block Block) *PreprepareMessageContentBuilder
	CreateViewChangeMessageContentBuilder(blockHeight BlockHeight, view View, preparedMessages *PreparedMessages) *ViewChangeMessageContentBuilder
	CreateNewViewMessageContentBuilder(blockHeight BlockHeight, view View, ppmcb *PreprepareMessageContentBuilder, confirmations []*ViewChangeMessageContentBuilder) *NewViewMessageContentBuilder
}

type HasMessageType interface {
	MessageType() MessageType
}

type Serializable interface {
	String() string
	Raw() []byte
}

type MessageTransporter interface {
	HasMessageType
	Sender() *SenderSignature
}

type MessageContent interface {
	HasMessageType
	Serializable

	Sender() *SenderSignature
}

// PP
type PreprepareMessage interface {
	MessageContent
	SignedHeader() *BlockRef
	Block() Block
}

type PrepareMessage interface {
	MessageContent
	SignedHeader() *BlockRef
}

type CommitMessage interface {
	MessageContent
	SignedHeader() *BlockRef
}

type ViewChangeMessage interface {
	MessageContent
	SignedHeader() *ViewChangeHeader
	Block() Block
}

type NewViewMessage interface {
	MessageContent
	SignedHeader() *NewViewHeader
	PreprepareMessageContent() *PreprepareMessageContent
	Block() Block
}

type PreprepareMessageImpl struct {
	Content *PreprepareMessageContent
	MyBlock Block
}

func (ppm *PreprepareMessageImpl) String() string {
	return ppm.Content.String()
}

func (ppm *PreprepareMessageImpl) MessageType() MessageType {
	return LEAN_HELIX_PREPREPARE
}

func (ppm *PreprepareMessageImpl) SignedHeader() *BlockRef {
	return ppm.Content.SignedHeader()
}

func (ppm *PreprepareMessageImpl) Sender() *SenderSignature {
	return ppm.Content.Sender()
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

func (pm *PrepareMessageImpl) String() string {
	return pm.Content.String()
}

func (pm *PrepareMessageImpl) MessageType() MessageType {
	return LEAN_HELIX_PREPARE
}

func (pm *PrepareMessageImpl) SignedHeader() *BlockRef {
	return pm.Content.SignedHeader()
}

func (pm *PrepareMessageImpl) Sender() *SenderSignature {
	return pm.Content.Sender()
}

func (pm *PrepareMessageImpl) Raw() []byte {
	return pm.Content.Raw()
}

type CommitMessageImpl struct {
	Content *CommitMessageContent
}

func (cm *CommitMessageImpl) String() string {
	return cm.Content.String()
}

func (cm *CommitMessageImpl) MessageType() MessageType {
	return LEAN_HELIX_COMMIT
}

func (cm *CommitMessageImpl) SignedHeader() *BlockRef {
	return cm.Content.SignedHeader()
}

func (cm *CommitMessageImpl) Sender() *SenderSignature {
	return cm.Content.Sender()
}

func (cm *CommitMessageImpl) Raw() []byte {
	return cm.Content.Raw()
}

type ViewChangeMessageImpl struct {
	Content *ViewChangeMessageContent
	MyBlock Block
}

func (vcm *ViewChangeMessageImpl) String() string {
	return vcm.Content.String()
}

func (vcm *ViewChangeMessageImpl) Block() Block {
	return vcm.MyBlock
}

func (vcm *ViewChangeMessageImpl) MessageType() MessageType {
	return LEAN_HELIX_VIEW_CHANGE
}

func (vcm *ViewChangeMessageImpl) SignedHeader() *ViewChangeHeader {
	return vcm.Content.SignedHeader()
}

func (vcm *ViewChangeMessageImpl) Sender() *SenderSignature {
	return vcm.Content.Sender()
}

func (vcm *ViewChangeMessageImpl) Raw() []byte {
	return vcm.Content.Raw()
}

type NewViewMessageImpl struct {
	Content *NewViewMessageContent
	MyBlock Block
}

func (nvm *NewViewMessageImpl) String() string {
	return nvm.Content.String()
}

func (nvm *NewViewMessageImpl) MessageType() MessageType {
	return LEAN_HELIX_NEW_VIEW
}

func (nvm *NewViewMessageImpl) SignedHeader() *NewViewHeader {
	return nvm.Content.SignedHeader()
}

func (nvm *NewViewMessageImpl) Sender() *SenderSignature {
	return nvm.Content.Sender()
}

func (nvm *NewViewMessageImpl) PreprepareMessageContent() *PreprepareMessageContent {
	return nvm.Content.PreprepareMessageContent()
}

func (nvm *NewViewMessageImpl) Block() Block {
	return nvm.MyBlock
}

func (nvm *NewViewMessageImpl) Raw() []byte {
	return nvm.Content.Raw()
}

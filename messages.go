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

// SHARED interfaces //
type Serializable interface {
	String() string
	Raw() []byte
}

type ConsensusMessage interface {
	Serializable
	MessageType() MessageType
	SenderPublicKey() Ed25519PublicKey
	BlockHeight() BlockHeight
}

/***************************************************/
/*            CORE Consensus MESSAGES              */
/***************************************************/

type PreprepareMessage interface {
	ConsensusMessage
	Content() *BlockRefContent
	Block() Block
}

type PrepareMessage interface {
	ConsensusMessage
	Content() *BlockRefContent
}

type CommitMessage interface {
	ConsensusMessage
	Content() *BlockRefContent
}

type ViewChangeMessage interface {
	ConsensusMessage
	Content() *ViewChangeMessageContent
	Block() Block
}

type NewViewMessage interface {
	ConsensusMessage
	Content() *NewViewMessageContent
	Block() Block
}

/***************************************************/
/*                 IMPLEMENTATIONS                 */
/***************************************************/

//------------
// Preprepare
//------------

type PreprepareMessageImpl struct {
	Content *PreprepareMessageContent
	block   Block
}

func (ppm *PreprepareMessageImpl) Raw() []byte {
	return ppm.Content.Raw()
}

func (ppm *PreprepareMessageImpl) String() string {
	return ppm.Content.String()
}

func (ppm *PreprepareMessageImpl) Block() Block {
	return ppm.block
}

func (ppm *PreprepareMessageImpl) MessageType() MessageType {
	return ppm.Content.SignedHeader().MessageType()
}

func (ppm *PreprepareMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return ppm.Content.Sender().SenderPublicKey()
}

func (ppm *PreprepareMessageImpl) BlockHeight() BlockHeight {
	return ppm.Content.SignedHeader().BlockHeight()
}

//---------
// Prepare
//---------
type PrepareMessageImpl struct {
	Content *PrepareMessageContent
}

func (pm *PrepareMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return pm.Content.Sender().SenderPublicKey()
}

func (pm *PrepareMessageImpl) BlockHeight() BlockHeight {
	return pm.Content.SignedHeader().BlockHeight()
}

func (pm *PrepareMessageImpl) Raw() []byte {
	return pm.Content.Raw()
}

func (pm *PrepareMessageImpl) String() string {
	return pm.Content.String()
}

func (pm *PrepareMessageImpl) MessageType() MessageType {
	return pm.Content.SignedHeader().MessageType()
}

func (pm *PrepareMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return pm.Content.Sender().SenderPublicKey()
}

func (pm *PrepareMessageImpl) BlockHeight() BlockHeight {
	return pm.Content.SignedHeader().BlockHeight()
}

//---------
// Commit
//---------
type CommitMessageImpl struct {
	Content *CommitMessageContent
}

func (cm *CommitMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return cm.Content.Sender().SenderPublicKey()
}

func (cm *CommitMessageImpl) BlockHeight() BlockHeight {
	return cm.Content.SignedHeader().BlockHeight()
}

func (cm *CommitMessageImpl) Raw() []byte {
	return cm.Content.Raw()
}

func (cm *CommitMessageImpl) String() string {
	return cm.Content.String()
}

func (cm *CommitMessageImpl) MessageType() MessageType {
	return cm.Content.SignedHeader().MessageType()
}

func (cm *CommitMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return cm.Content.Sender().SenderPublicKey()
}

func (cm *CommitMessageImpl) BlockHeight() BlockHeight {
	return cm.Content.SignedHeader().BlockHeight()
}

//-------------
// View Change
//-------------
type ViewChangeMessageImpl struct {
	Content *ViewChangeMessageContent
	block   Block
}

func (vcm *ViewChangeMessageImpl) Raw() []byte {
	return vcm.Content.Raw()
}

func (vcm *ViewChangeMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return vcm.Content.Sender().SenderPublicKey()
}

func (vcm *ViewChangeMessageImpl) BlockHeight() BlockHeight {
	return vcm.Content.SignedHeader().BlockHeight()
}

func (vcm *ViewChangeMessageImpl) String() string {
	return vcm.Content.String()
}

func (vcm *ViewChangeMessageImpl) Block() Block {
	return vcm.block
}

func (vcm *ViewChangeMessageImpl) MessageType() MessageType {
	return vcm.Content.SignedHeader().MessageType()
}

func (vcm *ViewChangeMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return vcm.Content.Sender().SenderPublicKey()
}

func (vcm *ViewChangeMessageImpl) BlockHeight() BlockHeight {
	return vcm.Content.SignedHeader().BlockHeight()
}

//----------
// New View
//----------
type NewViewMessageImpl struct {
	Content *NewViewMessageContent
	block   Block
}

func (nvm *NewViewMessageImpl) Raw() []byte {
	return nvm.Content.Raw()
}

func (nvm *NewViewMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return nvm.Content.Sender().SenderPublicKey()
}

func (nvm *NewViewMessageImpl) BlockHeight() BlockHeight {
	return nvm.Content.SignedHeader().BlockHeight()
}

func (nvm *NewViewMessageImpl) String() string {
	return nvm.Content.String()
}

func (nvm *NewViewMessageImpl) Block() Block {
	return nvm.block
}

func (nvm *NewViewMessageImpl) MessageType() MessageType {
	return nvm.Content.SignedHeader().MessageType()
}

func (nvm *NewViewMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return nvm.Content.Sender().SenderPublicKey()
}

func (nvm *NewViewMessageImpl) BlockHeight() BlockHeight {
	return nvm.Content.SignedHeader().BlockHeight()
}

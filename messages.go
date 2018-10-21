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
	CreateNewViewMessage(blockHeight BlockHeight, view View, preprepareContentBuilder *PreprepareContentBuilder, confirmations []*ViewChangeMessageContentBuilder, block Block) NewViewMessage

	// Helper methods
	CreatePreprepareMessageContentBuilder(blockHeight BlockHeight, view View, block Block) *PreprepareContentBuilder
	CreateViewChangeMessageContentBuilder(blockHeight BlockHeight, view View, preparedMessages *PreparedMessages) *ViewChangeMessageContentBuilder
	CreateNewViewMessageContentBuilder(blockHeight BlockHeight, view View, blockRefBuilder *PreprepareContentBuilder, confirmations []*ViewChangeMessageContentBuilder) *NewViewMessageContentBuilder
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
	View() View
}

/***************************************************/
/*            CORE Consensus MESSAGES              */
/***************************************************/

type PreprepareMessage interface {
	ConsensusMessage
	Content() *PreprepareContent
	Block() Block
}

type PrepareMessage interface {
	ConsensusMessage
	Content() *PrepareContent
}

type CommitMessage interface {
	ConsensusMessage
	Content() *CommitContent
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
	content *PreprepareContent
	block   Block
}

func (ppm *PreprepareMessageImpl) MessageType() MessageType {
	return ppm.content.SignedHeader().MessageType()
}

func (ppm *PreprepareMessageImpl) Content() *PreprepareContent {
	return ppm.content
}

func (ppm *PreprepareMessageImpl) Raw() []byte {
	return ppm.content.Raw()
}

func (ppm *PreprepareMessageImpl) String() string {
	return ppm.content.String()
}

func (ppm *PreprepareMessageImpl) Block() Block {
	return ppm.block
}

func (ppm *PreprepareMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return ppm.content.Sender().SenderPublicKey()
}

func (ppm *PreprepareMessageImpl) BlockHeight() BlockHeight {
	return ppm.content.SignedHeader().BlockHeight()
}

func (ppm *PreprepareMessageImpl) View() View {
	return ppm.content.SignedHeader().View()
}

//---------
// Prepare
//---------
type PrepareMessageImpl struct {
	content *PrepareContent
}

func (pm *PrepareMessageImpl) MessageType() MessageType {
	return pm.content.SignedHeader().MessageType()
}

func (pm *PrepareMessageImpl) Content() *PrepareContent {
	return pm.content
}

func (pm *PrepareMessageImpl) Raw() []byte {
	return pm.content.Raw()
}

func (pm *PrepareMessageImpl) String() string {
	return pm.content.String()
}

func (pm *PrepareMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return pm.content.Sender().SenderPublicKey()
}

func (pm *PrepareMessageImpl) BlockHeight() BlockHeight {
	return pm.content.SignedHeader().BlockHeight()
}
func (pm *PrepareMessageImpl) View() View {
	return pm.content.SignedHeader().View()
}

//---------
// Commit
//---------
type CommitMessageImpl struct {
	content *CommitContent
}

func (cm *CommitMessageImpl) MessageType() MessageType {
	return cm.content.SignedHeader().MessageType()
}

func (cm *CommitMessageImpl) Content() *CommitContent {
	return cm.content
}

func (cm *CommitMessageImpl) Raw() []byte {
	return cm.content.Raw()
}

func (cm *CommitMessageImpl) String() string {
	return cm.content.String()
}

func (cm *CommitMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return cm.content.Sender().SenderPublicKey()
}

func (cm *CommitMessageImpl) BlockHeight() BlockHeight {
	return cm.content.SignedHeader().BlockHeight()
}
func (cm *CommitMessageImpl) View() View {
	return cm.content.SignedHeader().View()
}

//-------------
// View Change
//-------------
type ViewChangeMessageImpl struct {
	content *ViewChangeMessageContent
	block   Block
}

func (vcm *ViewChangeMessageImpl) MessageType() MessageType {
	return vcm.content.SignedHeader().MessageType()
}

func (vcm *ViewChangeMessageImpl) Content() *ViewChangeMessageContent {
	return vcm.content
}

func (vcm *ViewChangeMessageImpl) Raw() []byte {
	return vcm.content.Raw()
}

func (vcm *ViewChangeMessageImpl) String() string {
	return vcm.content.String()
}

func (vcm *ViewChangeMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return vcm.content.Sender().SenderPublicKey()
}

func (vcm *ViewChangeMessageImpl) BlockHeight() BlockHeight {
	return vcm.content.SignedHeader().BlockHeight()
}

func (vcm *ViewChangeMessageImpl) Block() Block {
	return vcm.block
}
func (vcm *ViewChangeMessageImpl) View() View {
	return vcm.content.SignedHeader().View()
}

//----------
// New View
//----------
type NewViewMessageImpl struct {
	content *NewViewMessageContent
	block   Block
}

func (nvm *NewViewMessageImpl) MessageType() MessageType {
	return nvm.content.SignedHeader().MessageType()
}

func (nvm *NewViewMessageImpl) Content() *NewViewMessageContent {
	return nvm.content
}

func (nvm *NewViewMessageImpl) Raw() []byte {
	return nvm.content.Raw()
}

func (nvm *NewViewMessageImpl) String() string {
	return nvm.content.String()
}

func (nvm *NewViewMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return nvm.content.Sender().SenderPublicKey()
}

func (nvm *NewViewMessageImpl) BlockHeight() BlockHeight {
	return nvm.content.SignedHeader().BlockHeight()
}

func (nvm *NewViewMessageImpl) Block() Block {
	return nvm.block
}
func (nvm *NewViewMessageImpl) View() View {
	return nvm.content.SignedHeader().View()
}

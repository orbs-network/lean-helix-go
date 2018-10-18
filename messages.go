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
	CreateNewViewMessage(blockHeight BlockHeight, view View, blockRefContentBuilder *BlockRefContentBuilder, confirmations []*ViewChangeMessageContentBuilder, block Block) NewViewMessage

	// Helper methods
	CreatePreprepareMessageContentBuilder(blockHeight BlockHeight, view View, block Block) *BlockRefContentBuilder
	CreateViewChangeMessageContentBuilder(blockHeight BlockHeight, view View, preparedMessages *PreparedMessages) *ViewChangeMessageContentBuilder
	CreateNewViewMessageContentBuilder(blockHeight BlockHeight, view View, blockRefBuilder *BlockRefContentBuilder, confirmations []*ViewChangeMessageContentBuilder) *NewViewMessageContentBuilder
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
	MyContent *BlockRefContent
	MyBlock   Block
}

func (ppm *PreprepareMessageImpl) MessageType() MessageType {
	return ppm.MyContent.SignedHeader().MessageType()
}

func (ppm *PreprepareMessageImpl) Content() *BlockRefContent {
	return ppm.MyContent
}

func (ppm *PreprepareMessageImpl) Raw() []byte {
	return ppm.MyContent.Raw()
}

func (ppm *PreprepareMessageImpl) String() string {
	return ppm.MyContent.String()
}

func (ppm *PreprepareMessageImpl) Block() Block {
	return ppm.MyBlock
}

func (ppm *PreprepareMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return ppm.MyContent.Sender().SenderPublicKey()
}

func (ppm *PreprepareMessageImpl) BlockHeight() BlockHeight {
	return ppm.MyContent.SignedHeader().BlockHeight()
}

func (ppm *PreprepareMessageImpl) View() View {
	return ppm.MyContent.SignedHeader().View()
}

//---------
// Prepare
//---------
type PrepareMessageImpl struct {
	MyContent *BlockRefContent
}

func (pm *PrepareMessageImpl) MessageType() MessageType {
	return pm.MyContent.SignedHeader().MessageType()
}

func (pm *PrepareMessageImpl) Content() *BlockRefContent {
	return pm.MyContent
}

func (pm *PrepareMessageImpl) Raw() []byte {
	return pm.MyContent.Raw()
}

func (pm *PrepareMessageImpl) String() string {
	return pm.MyContent.String()
}

func (pm *PrepareMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return pm.MyContent.Sender().SenderPublicKey()
}

func (pm *PrepareMessageImpl) BlockHeight() BlockHeight {
	return pm.MyContent.SignedHeader().BlockHeight()
}
func (pm *PrepareMessageImpl) View() View {
	return pm.MyContent.SignedHeader().View()
}

//---------
// Commit
//---------
type CommitMessageImpl struct {
	MyContent *BlockRefContent
}

func (cm *CommitMessageImpl) MessageType() MessageType {
	return cm.MyContent.SignedHeader().MessageType()
}

func (cm *CommitMessageImpl) Content() *BlockRefContent {
	return cm.MyContent
}

func (cm *CommitMessageImpl) Raw() []byte {
	return cm.MyContent.Raw()
}

func (cm *CommitMessageImpl) String() string {
	return cm.MyContent.String()
}

func (cm *CommitMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return cm.MyContent.Sender().SenderPublicKey()
}

func (cm *CommitMessageImpl) BlockHeight() BlockHeight {
	return cm.MyContent.SignedHeader().BlockHeight()
}
func (cm *CommitMessageImpl) View() View {
	return cm.MyContent.SignedHeader().View()
}

//-------------
// View Change
//-------------
type ViewChangeMessageImpl struct {
	MyContent *ViewChangeMessageContent
	MyBlock   Block
}

func (vcm *ViewChangeMessageImpl) MessageType() MessageType {
	return vcm.MyContent.SignedHeader().MessageType()
}

func (vcm *ViewChangeMessageImpl) Content() *ViewChangeMessageContent {
	return vcm.MyContent
}

func (vcm *ViewChangeMessageImpl) Raw() []byte {
	return vcm.MyContent.Raw()
}

func (vcm *ViewChangeMessageImpl) String() string {
	return vcm.MyContent.String()
}

func (vcm *ViewChangeMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return vcm.MyContent.Sender().SenderPublicKey()
}

func (vcm *ViewChangeMessageImpl) BlockHeight() BlockHeight {
	return vcm.MyContent.SignedHeader().BlockHeight()
}

func (vcm *ViewChangeMessageImpl) Block() Block {
	return vcm.MyBlock
}
func (vcm *ViewChangeMessageImpl) View() View {
	return vcm.MyContent.SignedHeader().View()
}

//----------
// New View
//----------
type NewViewMessageImpl struct {
	MyContent *NewViewMessageContent
	MyBlock   Block
}

func (nvm *NewViewMessageImpl) MessageType() MessageType {
	return nvm.MyContent.SignedHeader().MessageType()
}

func (nvm *NewViewMessageImpl) Content() *NewViewMessageContent {
	return nvm.MyContent
}

func (nvm *NewViewMessageImpl) Raw() []byte {
	return nvm.MyContent.Raw()
}

func (nvm *NewViewMessageImpl) String() string {
	return nvm.MyContent.String()
}

func (nvm *NewViewMessageImpl) SenderPublicKey() Ed25519PublicKey {
	return nvm.MyContent.Sender().SenderPublicKey()
}

func (nvm *NewViewMessageImpl) BlockHeight() BlockHeight {
	return nvm.MyContent.SignedHeader().BlockHeight()
}

func (nvm *NewViewMessageImpl) Block() Block {
	return nvm.MyBlock
}
func (nvm *NewViewMessageImpl) View() View {
	return nvm.MyContent.SignedHeader().View()
}

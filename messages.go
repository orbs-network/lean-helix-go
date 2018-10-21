package leanhelix

import (
	. "github.com/orbs-network/lean-helix-go/primitives"
)

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

//------------
// Preprepare
//------------

type PreprepareMessage struct {
	ConsensusMessage
	content *PreprepareContent
	block   Block
}

func (ppm *PreprepareMessage) MessageType() MessageType {
	return ppm.content.SignedHeader().MessageType()
}

func (ppm *PreprepareMessage) Content() *PreprepareContent {
	return ppm.content
}

func (ppm *PreprepareMessage) Raw() []byte {
	return ppm.content.Raw()
}

func (ppm *PreprepareMessage) String() string {
	return ppm.content.String()
}

func (ppm *PreprepareMessage) Block() Block {
	return ppm.block
}

func (ppm *PreprepareMessage) SenderPublicKey() Ed25519PublicKey {
	return ppm.content.Sender().SenderPublicKey()
}

func (ppm *PreprepareMessage) BlockHeight() BlockHeight {
	return ppm.content.SignedHeader().BlockHeight()
}

func (ppm *PreprepareMessage) View() View {
	return ppm.content.SignedHeader().View()
}

func NewPreprepareMessage(content *PreprepareContent, block Block) *PreprepareMessage {
	return &PreprepareMessage{
		content: content,
		block:   block,
	}
}

//---------
// Prepare
//---------
type PrepareMessage struct {
	content *PrepareContent
}

func (pm *PrepareMessage) MessageType() MessageType {
	return pm.content.SignedHeader().MessageType()
}

func (pm *PrepareMessage) Content() *PrepareContent {
	return pm.content
}

func (pm *PrepareMessage) Raw() []byte {
	return pm.content.Raw()
}

func (pm *PrepareMessage) String() string {
	return pm.content.String()
}

func (pm *PrepareMessage) SenderPublicKey() Ed25519PublicKey {
	return pm.content.Sender().SenderPublicKey()
}

func (pm *PrepareMessage) BlockHeight() BlockHeight {
	return pm.content.SignedHeader().BlockHeight()
}
func (pm *PrepareMessage) View() View {
	return pm.content.SignedHeader().View()
}

func NewPrepareMessage(content *PrepareContent) *PrepareMessage {
	return &PrepareMessage{content: content}
}

//---------
// Commit
//---------
type CommitMessage struct {
	content *CommitContent
}

func (cm *CommitMessage) MessageType() MessageType {
	return cm.content.SignedHeader().MessageType()
}

func (cm *CommitMessage) Content() *CommitContent {
	return cm.content
}

func (cm *CommitMessage) Raw() []byte {
	return cm.content.Raw()
}

func (cm *CommitMessage) String() string {
	return cm.content.String()
}

func (cm *CommitMessage) SenderPublicKey() Ed25519PublicKey {
	return cm.content.Sender().SenderPublicKey()
}

func (cm *CommitMessage) BlockHeight() BlockHeight {
	return cm.content.SignedHeader().BlockHeight()
}
func (cm *CommitMessage) View() View {
	return cm.content.SignedHeader().View()
}

func NewCommitMessage(content *CommitContent) *CommitMessage {
	return &CommitMessage{content: content}
}

//-------------
// View Change
//-------------
type ViewChangeMessage struct {
	content *ViewChangeMessageContent
	block   Block
}

func (vcm *ViewChangeMessage) MessageType() MessageType {
	return vcm.content.SignedHeader().MessageType()
}

func (vcm *ViewChangeMessage) Content() *ViewChangeMessageContent {
	return vcm.content
}

func (vcm *ViewChangeMessage) Raw() []byte {
	return vcm.content.Raw()
}

func (vcm *ViewChangeMessage) String() string {
	return vcm.content.String()
}

func (vcm *ViewChangeMessage) SenderPublicKey() Ed25519PublicKey {
	return vcm.content.Sender().SenderPublicKey()
}

func (vcm *ViewChangeMessage) BlockHeight() BlockHeight {
	return vcm.content.SignedHeader().BlockHeight()
}

func (vcm *ViewChangeMessage) Block() Block {
	return vcm.block
}
func (vcm *ViewChangeMessage) View() View {
	return vcm.content.SignedHeader().View()
}

func NewViewChangeMessage(content *ViewChangeMessageContent, block Block) *ViewChangeMessage {
	return &ViewChangeMessage{
		content: content,
		block:   block,
	}
}

//----------
// New View
//----------
type NewViewMessage struct {
	content *NewViewMessageContent
	block   Block
}

func (nvm *NewViewMessage) MessageType() MessageType {
	return nvm.content.SignedHeader().MessageType()
}

func (nvm *NewViewMessage) Content() *NewViewMessageContent {
	return nvm.content
}

func (nvm *NewViewMessage) Raw() []byte {
	return nvm.content.Raw()
}

func (nvm *NewViewMessage) String() string {
	return nvm.content.String()
}

func (nvm *NewViewMessage) SenderPublicKey() Ed25519PublicKey {
	return nvm.content.Sender().SenderPublicKey()
}

func (nvm *NewViewMessage) BlockHeight() BlockHeight {
	return nvm.content.SignedHeader().BlockHeight()
}

func (nvm *NewViewMessage) Block() Block {
	return nvm.block
}
func (nvm *NewViewMessage) View() View {
	return nvm.content.SignedHeader().View()
}

func NewNewViewMessage(content *NewViewMessageContent, block Block) *NewViewMessage {
	return &NewViewMessage{
		content: content,
		block:   block,
	}
}

package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

// SHARED interfaces //
type Serializable interface {
	String() string
	Raw() []byte
}

type ConsensusRawMessage interface {
	Content() []byte
	Block() Block
	ToConsensusMessage() ConsensusMessage
}

type ConsensusRawMessageConverter interface {
	ToConsensusRawMessage() ConsensusRawMessage
}

type ConsensusMessage interface {
	Serializable
	ConsensusRawMessageConverter
	MessageType() protocol.MessageType
	SenderPublicKey() primitives.MemberId
	BlockHeight() primitives.BlockHeight
	View() primitives.View
}

/***************************************************/
/*            CORE Consensus MESSAGES              */
/***************************************************/

//------------
// Preprepare
//------------

type PreprepareMessage struct {
	ConsensusMessage
	content *protocol.PreprepareContent
	block   Block
}

func (ppm *PreprepareMessage) MessageType() protocol.MessageType {
	return ppm.content.SignedHeader().MessageType()
}

func (ppm *PreprepareMessage) Content() *protocol.PreprepareContent {
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

func (ppm *PreprepareMessage) SenderPublicKey() primitives.MemberId {
	return ppm.content.Sender().MemberId()
}

func (ppm *PreprepareMessage) BlockHeight() primitives.BlockHeight {
	return ppm.content.SignedHeader().BlockHeight()
}

func (ppm *PreprepareMessage) View() primitives.View {
	return ppm.content.SignedHeader().View()
}

func (ppm *PreprepareMessage) ToConsensusRawMessage() ConsensusRawMessage {
	return CreateConsensusRawMessage(ppm.content.Raw(), ppm.block)
}

func NewPreprepareMessage(content *protocol.PreprepareContent, block Block) *PreprepareMessage {
	return &PreprepareMessage{
		content: content,
		block:   block,
	}
}

//---------
// Prepare
//---------
type PrepareMessage struct {
	content *protocol.PrepareContent
}

func (pm *PrepareMessage) MessageType() protocol.MessageType {
	return pm.content.SignedHeader().MessageType()
}

func (pm *PrepareMessage) Content() *protocol.PrepareContent {
	return pm.content
}

func (pm *PrepareMessage) Raw() []byte {
	return pm.content.Raw()
}

func (pm *PrepareMessage) String() string {
	return pm.content.String()
}

func (pm *PrepareMessage) SenderPublicKey() primitives.MemberId {
	return pm.content.Sender().MemberId()
}

func (pm *PrepareMessage) BlockHeight() primitives.BlockHeight {
	return pm.content.SignedHeader().BlockHeight()
}
func (pm *PrepareMessage) View() primitives.View {
	return pm.content.SignedHeader().View()
}

func (pm *PrepareMessage) ToConsensusRawMessage() ConsensusRawMessage {
	return CreateConsensusRawMessage(pm.content.Raw(), nil)
}

func NewPrepareMessage(content *protocol.PrepareContent) *PrepareMessage {
	return &PrepareMessage{content: content}
}

//---------
// Commit
//---------
type CommitMessage struct {
	content *protocol.CommitContent
}

func (cm *CommitMessage) MessageType() protocol.MessageType {
	return cm.content.SignedHeader().MessageType()
}

func (cm *CommitMessage) Content() *protocol.CommitContent {
	return cm.content
}

func (cm *CommitMessage) Raw() []byte {
	return cm.content.Raw()
}

func (cm *CommitMessage) String() string {
	return cm.content.String()
}

func (cm *CommitMessage) SenderPublicKey() primitives.MemberId {
	return cm.content.Sender().MemberId()
}

func (cm *CommitMessage) BlockHeight() primitives.BlockHeight {
	return cm.content.SignedHeader().BlockHeight()
}
func (cm *CommitMessage) View() primitives.View {
	return cm.content.SignedHeader().View()
}

func (cm *CommitMessage) ToConsensusRawMessage() ConsensusRawMessage {
	return CreateConsensusRawMessage(cm.content.Raw(), nil)
}

func NewCommitMessage(content *protocol.CommitContent) *CommitMessage {
	return &CommitMessage{content: content}
}

//-------------
// View Change
//-------------
type ViewChangeMessage struct {
	content *protocol.ViewChangeMessageContent
	block   Block
}

func (vcm *ViewChangeMessage) MessageType() protocol.MessageType {
	return vcm.content.SignedHeader().MessageType()
}

func (vcm *ViewChangeMessage) Content() *protocol.ViewChangeMessageContent {
	return vcm.content
}

func (vcm *ViewChangeMessage) Raw() []byte {
	return vcm.content.Raw()
}

func (vcm *ViewChangeMessage) String() string {
	return vcm.content.String()
}

func (vcm *ViewChangeMessage) SenderPublicKey() primitives.MemberId {
	return vcm.content.Sender().MemberId()
}

func (vcm *ViewChangeMessage) BlockHeight() primitives.BlockHeight {
	return vcm.content.SignedHeader().BlockHeight()
}

func (vcm *ViewChangeMessage) Block() Block {
	return vcm.block
}
func (vcm *ViewChangeMessage) View() primitives.View {
	return vcm.content.SignedHeader().View()
}

func (vcm *ViewChangeMessage) ToConsensusRawMessage() ConsensusRawMessage {
	return CreateConsensusRawMessage(vcm.content.Raw(), vcm.block)
}

func NewViewChangeMessage(content *protocol.ViewChangeMessageContent, block Block) *ViewChangeMessage {
	return &ViewChangeMessage{
		content: content,
		block:   block,
	}
}

//----------
// New View
//----------
type NewViewMessage struct {
	content *protocol.NewViewMessageContent
	block   Block
}

func (nvm *NewViewMessage) MessageType() protocol.MessageType {
	return nvm.content.SignedHeader().MessageType()
}

func (nvm *NewViewMessage) Content() *protocol.NewViewMessageContent {
	return nvm.content
}

func (nvm *NewViewMessage) Raw() []byte {
	return nvm.content.Raw()
}

func (nvm *NewViewMessage) String() string {
	return nvm.content.String()
}

func (nvm *NewViewMessage) SenderPublicKey() primitives.MemberId {
	return nvm.content.Sender().MemberId()
}

func (nvm *NewViewMessage) BlockHeight() primitives.BlockHeight {
	return nvm.content.SignedHeader().BlockHeight()
}

func (nvm *NewViewMessage) Block() Block {
	return nvm.block
}
func (nvm *NewViewMessage) View() primitives.View {
	return nvm.content.SignedHeader().View()
}

func (nvm *NewViewMessage) ToConsensusRawMessage() ConsensusRawMessage {
	return CreateConsensusRawMessage(nvm.content.Raw(), nvm.block)
}

func NewNewViewMessage(content *protocol.NewViewMessageContent, block Block) *NewViewMessage {
	return &NewViewMessage{
		content: content,
		block:   block,
	}
}

func extractConfirmationsFromViewChangeMessages(vcms []*ViewChangeMessage) []*protocol.ViewChangeMessageContentBuilder {
	if len(vcms) == 0 {
		return nil
	}

	res := make([]*protocol.ViewChangeMessageContentBuilder, 0, len(vcms))
	for _, vcm := range vcms {
		header := vcm.content.SignedHeader()
		sender := vcm.content.Sender()
		proof := header.PreparedProof()
		var proofBuilder *protocol.PreparedProofBuilder = nil
		if proof != nil && len(proof.Raw()) > 0 {
			ppBlockRefBuilder := &protocol.BlockRefBuilder{
				MessageType: proof.PreprepareBlockRef().MessageType(),
				BlockHeight: proof.PreprepareBlockRef().BlockHeight(),
				View:        proof.PreprepareBlockRef().View(),
				BlockHash:   proof.PreprepareBlockRef().BlockHash(),
			}
			ppSender := &protocol.SenderSignatureBuilder{
				MemberId:  proof.PreprepareSender().MemberId(),
				Signature: proof.PreprepareSender().Signature(),
			}
			pBlockRef := &protocol.BlockRefBuilder{
				MessageType: proof.PrepareBlockRef().MessageType(),
				BlockHeight: proof.PrepareBlockRef().BlockHeight(),
				View:        proof.PrepareBlockRef().View(),
				BlockHash:   proof.PrepareBlockRef().BlockHash(),
			}
			pSendersIter := proof.PrepareSendersIterator()
			pSenders := make([]*protocol.SenderSignatureBuilder, 0, 1)

			for {
				if !pSendersIter.HasNext() {
					break
				}
				nextPSender := pSendersIter.NextPrepareSenders()
				pSender := &protocol.SenderSignatureBuilder{
					MemberId:  nextPSender.MemberId(),
					Signature: nextPSender.Signature(),
				}

				pSenders = append(pSenders, pSender)
			}

			proofBuilder = &protocol.PreparedProofBuilder{
				PreprepareBlockRef: ppBlockRefBuilder,
				PreprepareSender:   ppSender,
				PrepareBlockRef:    pBlockRef,
				PrepareSenders:     pSenders,
			}
		}

		viewChangeMessageContentBuilder := &protocol.ViewChangeMessageContentBuilder{
			SignedHeader: &protocol.ViewChangeHeaderBuilder{
				MessageType:   header.MessageType(),
				BlockHeight:   header.BlockHeight(),
				View:          header.View(),
				PreparedProof: proofBuilder,
			},
			Sender: &protocol.SenderSignatureBuilder{
				MemberId:  sender.MemberId(),
				Signature: sender.Signature(),
			},
		}
		res = append(res, viewChangeMessageContentBuilder)

	}
	return res
	//const viewChangeVotes: ViewChangeContent[] =
	//	viewChangeMessages.map(vc =>
	//		({ signedHeader: vc.content.signedHeader, sender: vc.content.sender }));
}

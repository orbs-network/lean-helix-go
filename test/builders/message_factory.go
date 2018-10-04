package builders

import lh "github.com/orbs-network/lean-helix-go"

// MessageFactory - see receivers below

type mockMessageFactory struct {
	CalculateBlockHash func(block lh.Block) lh.BlockHash
	keyManager         lh.KeyManager
	MyPK               lh.PublicKey
}

// signedHeader
type blockRef struct {
	messageType lh.MessageType
	height      lh.BlockHeight
	view        lh.View
	blockHash   lh.BlockHash
}

func (ref *blockRef) MessageType() lh.MessageType {
	return ref.messageType
}

func (ref *blockRef) BlockHeight() lh.BlockHeight {
	return ref.height
}

func (ref *blockRef) View() lh.View {
	return ref.view
}

func (ref *blockRef) BlockHash() lh.BlockHash {
	return ref.blockHash
}

// senderSignature
type senderSignature struct {
	senderPublicKey lh.PublicKey
	signature       lh.Signature
}

func (s *senderSignature) SenderPublicKey() lh.PublicKey {
	return s.senderPublicKey
}

func (s *senderSignature) Signature() lh.Signature {
	return s.signature
}

// PP
type preprepareMessage struct {
	signedHeader *blockRef
	sender       lh.SenderSignature
	block        lh.Block
}

func (ppm *preprepareMessage) SignedHeader() lh.BlockRef {
	return ppm.signedHeader
}

func (ppm *preprepareMessage) MessageType() lh.MessageType {
	return ppm.signedHeader.messageType
}

func (ppm *preprepareMessage) BlockHeight() lh.BlockHeight {
	return ppm.signedHeader.height
}

func (ppm *preprepareMessage) View() lh.View {
	return ppm.signedHeader.view
}

func (ppm *preprepareMessage) BlockHash() lh.BlockHash {
	return ppm.signedHeader.blockHash
}

func (ppm *preprepareMessage) Sender() lh.SenderSignature {
	return ppm.sender
}

func (ppm *preprepareMessage) Block() lh.Block {
	return ppm.block
}

// P
type prepareMessage struct {
	signedHeader *blockRef
	sender       lh.SenderSignature
	block        lh.Block
}

func (pm *prepareMessage) SignedHeader() lh.BlockRef {
	return pm.signedHeader
}

func (pm *prepareMessage) MessageType() lh.MessageType {
	return pm.signedHeader.messageType
}

func (pm *prepareMessage) BlockHeight() lh.BlockHeight {
	return pm.signedHeader.height
}

func (pm *prepareMessage) View() lh.View {
	return pm.signedHeader.view
}

func (pm *prepareMessage) BlockHash() lh.BlockHash {
	return pm.signedHeader.blockHash
}

func (pm *prepareMessage) Sender() lh.SenderSignature {
	return pm.sender
}

// C
type commitMessage struct {
	signedHeader *blockRef
	sender       lh.SenderSignature
}

func (cm *commitMessage) SignedHeader() lh.BlockRef {
	return cm.signedHeader
}

func (cm *commitMessage) MessageType() lh.MessageType {
	return cm.signedHeader.messageType
}

func (cm *commitMessage) BlockHeight() lh.BlockHeight {
	return cm.signedHeader.height
}

func (cm *commitMessage) View() lh.View {
	return cm.signedHeader.view
}

func (cm *commitMessage) BlockHash() lh.BlockHash {
	return cm.signedHeader.blockHash
}

func (cm *commitMessage) Sender() lh.SenderSignature {
	return cm.sender
}

// VC
type viewChangeMessage struct {
	signedHeader *viewChangeHeader
	sender       lh.SenderSignature
	block        lh.Block
}

func (vcm *viewChangeMessage) MessageType() lh.MessageType {
	return vcm.signedHeader.messageType
}

func (vcm *viewChangeMessage) BlockHeight() lh.BlockHeight {
	return vcm.signedHeader.height
}

func (vcm *viewChangeMessage) View() lh.View {
	return vcm.signedHeader.view
}

func (vcm *viewChangeMessage) SignedHeader() lh.ViewChangeHeader {
	return vcm.signedHeader
}

func (vcm *viewChangeMessage) Sender() lh.SenderSignature {
	return vcm.sender
}

func (vcm *viewChangeMessage) Block() lh.Block {
	return vcm.block
}

func (vcm *viewChangeMessage) PreparedProof() lh.PreparedProof {
	return vcm.signedHeader.preparedProof
}

// NV
type newViewMessage struct {
	blockRef                *blockRef // TODO doesn't need lh.BlockHash so maybe replace with  BlockHeight() and  View()
	viewChangeConfirmations []lh.ViewChangeConfirmation
}

func (nvm *newViewMessage) ViewChangeConfirmations() []lh.ViewChangeConfirmation {
	return nvm.viewChangeConfirmations
}

type viewChangeHeader struct {
	messageType   lh.MessageType
	height        lh.BlockHeight
	view          lh.View
	preparedProof *preparedProof
}

func (v *viewChangeHeader) MessageType() lh.MessageType {
	return v.messageType
}

func (v *viewChangeHeader) BlockHeight() lh.BlockHeight {
	return v.height
}

func (v *viewChangeHeader) View() lh.View {
	return v.view
}

func (v *viewChangeHeader) PreparedProof() lh.PreparedProof {
	return v.preparedProof
}

// Prepared Proof

type preparedProof struct {
	preprepare lh.PreprepareMessage
	prepares   []lh.PrepareMessage
}

func (pf *preparedProof) PreprepareMessage() lh.PreprepareMessage {
	return pf.preprepare
}

func (pf *preparedProof) PrepareMessages() []lh.PrepareMessage {
	return pf.prepares
}

// MessageFactory receivers
func NewMockMessageFactory(calculateBlockHash func(block lh.Block) lh.BlockHash, keyManager lh.KeyManager) *mockMessageFactory {
	return &mockMessageFactory{
		CalculateBlockHash: calculateBlockHash,
		keyManager:         keyManager,
		MyPK:               keyManager.MyPublicKey(),
	}
}

func (mf *mockMessageFactory) CreatePreprepareMessage(blockHeight lh.BlockHeight, view lh.View, block lh.Block) lh.PreprepareMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockRef := &blockRef{
		messageType: lh.MESSAGE_TYPE_PREPREPARE,
		height:      blockHeight,
		view:        view,
		blockHash:   blockHash,
	}

	sender := mf.keyManager.SignBlockRef(blockRef)

	result := &preprepareMessage{
		signedHeader: blockRef,
		sender:       sender,
		block:        block,
	}

	return result
}

func (mf *mockMessageFactory) CreatePrepareMessage(blockHeight lh.BlockHeight, view lh.View, block lh.Block) lh.PrepareMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockRef := &blockRef{
		messageType: lh.MESSAGE_TYPE_PREPARE,
		height:      blockHeight,
		view:        view,
		blockHash:   blockHash,
	}

	sender := mf.keyManager.SignBlockRef(blockRef)

	result := &prepareMessage{
		signedHeader: blockRef,
		sender:       sender,
		block:        block,
	}

	return result
}

func (mf *mockMessageFactory) CreateCommitMessage(blockHeight lh.BlockHeight, view lh.View, block lh.Block) lh.CommitMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockRef := &blockRef{
		messageType: lh.MESSAGE_TYPE_COMMIT,
		height:      blockHeight,
		view:        view,
		blockHash:   blockHash,
	}

	sender := mf.keyManager.SignBlockRef(blockRef)

	result := &commitMessage{
		signedHeader: blockRef,
		sender:       sender,
	}

	return result
}

func (mf *mockMessageFactory) CreateViewChangeMessage(blockHeight lh.BlockHeight, view lh.View, preprepare lh.PreprepareMessage, prepares []lh.PrepareMessage) lh.ViewChangeMessage {
	var (
		preparedProof *preparedProof
		block         lh.Block
	)
	if preprepare != nil && prepares != nil {
		preparedProof = generatePreparedProof(preprepare, prepares)
		block = preprepare.Block()
	}

	header := &viewChangeHeader{
		messageType:   lh.MESSAGE_TYPE_VIEW_CHANGE,
		height:        blockHeight,
		view:          view,
		preparedProof: preparedProof,
	}

	sender := mf.keyManager.SignViewChange(header)

	result := &viewChangeMessage{
		signedHeader: header,
		sender:       sender,
		block:        block,
	}

	return result
}

// TODO add CreateNewViewMessage

func generatePreparedProof(preprepare lh.PreprepareMessage, prepares []lh.PrepareMessage) *preparedProof {
	return &preparedProof{
		preprepare: preprepare,
		prepares:   prepares,
	}
}

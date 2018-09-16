package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
)

// MessageFactory - see receivers below
type MessageFactory interface {
	CreatePreprepareMessage(term lh.BlockHeight, view lh.ViewCounter, block lh.Block) lh.PreprepareMessage
	CreatePrepareMessage(term lh.BlockHeight, view lh.ViewCounter, block lh.Block) lh.PrepareMessage
	CreatePreparedProof(preprepare lh.PreprepareMessage, prepares []lh.PrepareMessage) lh.PreparedProof
}

type messageFactory struct {
	CalculateBlockHash func(block lh.Block) lh.BlockHash
	keyManager         lh.KeyManager
	MyPK               lh.PublicKey
}

// blockRef
type blockRef struct {
	messageType lh.MessageType
	term        lh.BlockHeight
	view        lh.ViewCounter
	blockHash   lh.BlockHash
}

func (ref *blockRef) MessageType() lh.MessageType {
	return ref.messageType
}

func (ref *blockRef) Term() lh.BlockHeight {
	return ref.term
}

func (ref *blockRef) View() lh.ViewCounter {
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
	blockRef *blockRef
	sender   lh.SenderSignature
	block    lh.Block
}

func (ppm *preprepareMessage) MessageType() lh.MessageType {
	return ppm.blockRef.messageType
}

func (ppm *preprepareMessage) Term() lh.BlockHeight {
	return ppm.blockRef.term
}

func (ppm *preprepareMessage) View() lh.ViewCounter {
	return ppm.blockRef.view
}

func (ppm *preprepareMessage) BlockHash() lh.BlockHash {
	return ppm.blockRef.blockHash
}

func (ppm *preprepareMessage) Sender() lh.SenderSignature {
	return ppm.sender
}

func (ppm *preprepareMessage) Block() lh.Block {
	return ppm.block
}

// P
type prepareMessage struct {
	blockRef *blockRef
	sender   lh.SenderSignature
	block    lh.Block
}

func (pm *prepareMessage) MessageType() lh.MessageType {
	return pm.blockRef.messageType
}

func (pm *prepareMessage) Term() lh.BlockHeight {
	return pm.blockRef.term
}

func (pm *prepareMessage) View() lh.ViewCounter {
	return pm.blockRef.view
}

func (pm *prepareMessage) BlockHash() lh.BlockHash {
	return pm.blockRef.blockHash
}

func (pm *prepareMessage) Sender() lh.SenderSignature {
	return pm.sender
}

// C
type commitMessage struct {
	blockRef *blockRef
	sender   lh.SenderSignature
}

func (cm *commitMessage) MessageType() lh.MessageType {
	return cm.blockRef.messageType
}

func (cm *commitMessage) Term() lh.BlockHeight {
	return cm.blockRef.term
}

func (cm *commitMessage) View() lh.ViewCounter {
	return cm.blockRef.view
}

func (cm *commitMessage) BlockHash() lh.BlockHash {
	return cm.blockRef.blockHash
}

func (cm *commitMessage) Sender() lh.SenderSignature {
	return cm.sender
}

// VC
type viewChangeMessage struct {
	blockRef      *blockRef
	sender        lh.SenderSignature
	block         lh.Block
	preparedProof lh.PreparedProof
}

func (vcm *viewChangeMessage) MessageType() lh.MessageType {
	return vcm.blockRef.messageType
}

func (vcm *viewChangeMessage) Term() lh.BlockHeight {
	return vcm.blockRef.term
}

func (vcm *viewChangeMessage) View() lh.ViewCounter {
	return vcm.blockRef.view
}

func (vcm *viewChangeMessage) BlockHash() lh.BlockHash {
	return vcm.blockRef.blockHash
}

func (vcm *viewChangeMessage) Sender() lh.SenderSignature {
	return vcm.sender
}

func (vcm *viewChangeMessage) Block() lh.Block {
	return vcm.block
}

func (vcm *viewChangeMessage) PreparedProof() lh.PreparedProof {
	return vcm.preparedProof
}

// NV
type newViewMessage struct {
	blockRef                *blockRef // TODO doesn't need BlockHash so maybe replace with Term() and View()
	viewChangeConfirmations []lh.ViewChangeConfirmation
}

func (nvm *newViewMessage) ViewChangeConfirmations() []lh.ViewChangeConfirmation {
	return nvm.viewChangeConfirmations
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
func NewMessageFactory(calculateBlockHash func(block lh.Block) lh.BlockHash, keyManager lh.KeyManager) *messageFactory {
	return &messageFactory{
		CalculateBlockHash: calculateBlockHash,
		keyManager:         keyManager,
		MyPK:               keyManager.MyID(),
	}
}

func (mf *messageFactory) CreatePreprepareMessage(term lh.BlockHeight, view lh.ViewCounter, block lh.Block) lh.PreprepareMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockRef := &blockRef{
		messageType: lh.MESSAGE_TYPE_PREPREPARE,
		term:        term,
		view:        view,
		blockHash:   blockHash,
	}

	sender := mf.keyManager.SignBlockRef(blockRef)

	result := &preprepareMessage{
		blockRef: blockRef,
		sender:   sender,
		block:    block,
	}

	return result
}

func (mf *messageFactory) CreatePrepareMessage(term lh.BlockHeight, view lh.ViewCounter, block lh.Block) lh.PrepareMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockRef := &blockRef{
		messageType: lh.MESSAGE_TYPE_PREPARE,
		term:        term,
		view:        view,
		blockHash:   blockHash,
	}

	sender := mf.keyManager.SignBlockRef(blockRef)

	result := &prepareMessage{
		blockRef: blockRef,
		sender:   sender,
		block:    block,
	}

	return result
}

func (mf *messageFactory) CreateCommitMessage(term lh.BlockHeight, view lh.ViewCounter, block lh.Block) lh.CommitMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockRef := &blockRef{
		messageType: lh.MESSAGE_TYPE_COMMIT,
		term:        term,
		view:        view,
		blockHash:   blockHash,
	}

	sender := mf.keyManager.SignBlockRef(blockRef)

	result := &commitMessage{
		blockRef: blockRef,
		sender:   sender,
	}

	return result
}

func (mf *messageFactory) CreateViewChangeMessage(term lh.BlockHeight, view lh.ViewCounter, preprepare lh.PreprepareMessage, prepares []lh.PrepareMessage) lh.ViewChangeMessage {
	var (
		preparedProof lh.PreparedProof
		block         lh.Block
		blockHash     lh.BlockHash
	)
	if preprepare != nil && prepares != nil {
		preparedProof = generatePreparedProof(preprepare, prepares)
		block = preprepare.Block()
		blockHash = mf.CalculateBlockHash(block)
	}

	blockRef := &blockRef{
		messageType: lh.MESSAGE_TYPE_VIEW_CHANGE,
		term:        term,
		view:        view,
		blockHash:   blockHash,
	}

	sender := mf.keyManager.SignBlockRef(blockRef)

	result := &viewChangeMessage{
		blockRef:      blockRef,
		sender:        sender,
		block:         block,
		preparedProof: preparedProof,
	}

	return result
}

func generatePreparedProof(preprepare lh.PreprepareMessage, prepares []lh.PrepareMessage) lh.PreparedProof {
	return &preparedProof{
		preprepare: preprepare,
		prepares:   prepares,
	}
}

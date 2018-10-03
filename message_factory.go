package leanhelix

// MessageFactory - see receivers below
type MessageFactory interface {
	CreatePreprepareMessage(blockHeight BlockHeight, view View, block Block) PreprepareMessage
	CreatePrepareMessage(blockHeight BlockHeight, view View, block Block) PrepareMessage
	//CreatePreparedProof(preprepare PreprepareMessage, prepares []PrepareMessage) PreparedProof
}

type messageFactory struct {
	CalculateBlockHash func(block Block) BlockHash
	keyManager         KeyManager
	MyPK               PublicKey
}

// blockRef
type blockRef struct {
	messageType MessageType
	height      BlockHeight
	view        View
	blockHash   BlockHash
}

func (ref *blockRef) MessageType() MessageType {
	return ref.messageType
}

func (ref *blockRef) BlockHeight() BlockHeight {
	return ref.height
}

func (ref *blockRef) View() View {
	return ref.view
}

func (ref *blockRef) BlockHash() BlockHash {
	return ref.blockHash
}

// senderSignature
type senderSignature struct {
	senderPublicKey PublicKey
	signature       Signature
}

func (s *senderSignature) SenderPublicKey() PublicKey {
	return s.senderPublicKey
}

func (s *senderSignature) Signature() Signature {
	return s.signature
}

// PP
type preprepareMessage struct {
	blockRef *blockRef
	sender   SenderSignature
	block    Block
}

func (ppm *preprepareMessage) MessageType() MessageType {
	return ppm.blockRef.messageType
}

func (ppm *preprepareMessage) BlockHeight() BlockHeight {
	return ppm.blockRef.height
}

func (ppm *preprepareMessage) View() View {
	return ppm.blockRef.view
}

func (ppm *preprepareMessage) BlockHash() BlockHash {
	return ppm.blockRef.blockHash
}

func (ppm *preprepareMessage) Sender() SenderSignature {
	return ppm.sender
}

func (ppm *preprepareMessage) Block() Block {
	return ppm.block
}

// P
type prepareMessage struct {
	blockRef *blockRef
	sender   SenderSignature
	block    Block
}

func (pm *prepareMessage) MessageType() MessageType {
	return pm.blockRef.messageType
}

func (pm *prepareMessage) BlockHeight() BlockHeight {
	return pm.blockRef.height
}

func (pm *prepareMessage) View() View {
	return pm.blockRef.view
}

func (pm *prepareMessage) BlockHash() BlockHash {
	return pm.blockRef.blockHash
}

func (pm *prepareMessage) Sender() SenderSignature {
	return pm.sender
}

// C
type commitMessage struct {
	blockRef *blockRef
	sender   SenderSignature
}

func (cm *commitMessage) MessageType() MessageType {
	return cm.blockRef.messageType
}

func (cm *commitMessage) BlockHeight() BlockHeight {
	return cm.blockRef.height
}

func (cm *commitMessage) View() View {
	return cm.blockRef.view
}

func (cm *commitMessage) BlockHash() BlockHash {
	return cm.blockRef.blockHash
}

func (cm *commitMessage) Sender() SenderSignature {
	return cm.sender
}

// VC
type viewChangeMessage struct {
	blockRef      *blockRef
	sender        SenderSignature
	block         Block
	preparedProof PreparedProof
}

func (vcm *viewChangeMessage) MessageType() MessageType {
	return vcm.blockRef.messageType
}

func (vcm *viewChangeMessage) BlockHeight() BlockHeight {
	return vcm.blockRef.height
}

func (vcm *viewChangeMessage) View() View {
	return vcm.blockRef.view
}

func (vcm *viewChangeMessage) BlockHash() BlockHash {
	return vcm.blockRef.blockHash
}

func (vcm *viewChangeMessage) Sender() SenderSignature {
	return vcm.sender
}

func (vcm *viewChangeMessage) Block() Block {
	return vcm.block
}

func (vcm *viewChangeMessage) PreparedProof() PreparedProof {
	return vcm.preparedProof
}

// NV
type newViewMessage struct {
	blockRef                *blockRef // TODO doesn't need BlockHash so maybe replace with BlockHeight() and View()
	viewChangeConfirmations []ViewChangeConfirmation
}

func (nvm *newViewMessage) ViewChangeConfirmations() []ViewChangeConfirmation {
	return nvm.viewChangeConfirmations
}

// Prepared Proof

type preparedProof struct {
	preprepare PreprepareMessage
	prepares   []PrepareMessage
}

func (pf *preparedProof) PreprepareMessage() PreprepareMessage {
	return pf.preprepare
}

func (pf *preparedProof) PrepareMessages() []PrepareMessage {
	return pf.prepares
}

// MessageFactory receivers
func NewMessageFactory(calculateBlockHash func(block Block) BlockHash, keyManager KeyManager) *messageFactory {
	return &messageFactory{
		CalculateBlockHash: calculateBlockHash,
		keyManager:         keyManager,
		MyPK:               keyManager.MyPublicKey(),
	}
}

func (mf *messageFactory) CreatePreprepareMessage(blockHeight BlockHeight, view View, block Block) PreprepareMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockRef := &blockRef{
		messageType: MESSAGE_TYPE_PREPREPARE,
		height:      blockHeight,
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

func (mf *messageFactory) CreatePrepareMessage(blockHeight BlockHeight, view View, block Block) PrepareMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockRef := &blockRef{
		messageType: MESSAGE_TYPE_PREPARE,
		height:      blockHeight,
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

func (mf *messageFactory) CreateCommitMessage(blockHeight BlockHeight, view View, block Block) CommitMessage {
	blockHash := mf.CalculateBlockHash(block)

	blockRef := &blockRef{
		messageType: MESSAGE_TYPE_COMMIT,
		height:      blockHeight,
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

func (mf *messageFactory) CreateViewChangeMessage(blockHeight BlockHeight, view View, preprepare PreprepareMessage, prepares []PrepareMessage) ViewChangeMessage {
	var (
		preparedProof PreparedProof
		block         Block
		blockHash     BlockHash
	)
	if preprepare != nil && prepares != nil {
		preparedProof = generatePreparedProof(preprepare, prepares)
		block = preprepare.Block()
		blockHash = mf.CalculateBlockHash(block)
	}

	blockRef := &blockRef{
		messageType: MESSAGE_TYPE_VIEW_CHANGE,
		height:      blockHeight,
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

func generatePreparedProof(preprepare PreprepareMessage, prepares []PrepareMessage) PreparedProof {
	return &preparedProof{
		preprepare: preprepare,
		prepares:   prepares,
	}
}

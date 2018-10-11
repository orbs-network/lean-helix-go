package test

import (
	"fmt"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"testing"
)

// Reference implementation of KeyManagerImpl usage

func (k *MockKeyManager) Sign(content []byte) []byte {
	return []byte{'e', 'n', 'i', 'g', 'm', 'a'} // Put something here
}

type MockMessageFactory struct {
	mock.Mock
}

type MockKeyManager struct {
	mock.Mock
}

func (k *MockKeyManager) Sign(content []byte) []byte {
	return []byte{'e', 'n', 'i', 'g', 'm', 'a'}
}

// Message creation methods

func (m *MockMessageFactory) CreatePreprepareMessage(blockRef leanhelix.BlockRef, sender leanhelix.SenderSignature, block leanhelix.Block) leanhelix.PreprepareMessage {
	panic("this is Orbs code")
}

func (m *MockMessageFactory) CreatePrepareMessage(blockRef leanhelix.BlockRef, sender leanhelix.SenderSignature) leanhelix.PrepareMessage {
	panic("this is Orbs code")
}

func (m *MockMessageFactory) CreateCommitMessage(blockRef leanhelix.BlockRef, sender leanhelix.SenderSignature) leanhelix.CommitMessage {
	panic("this is Orbs code")
}

func (m *MockMessageFactory) CreateViewChangeMessage(vcHeader leanhelix.ViewChangeHeader, sender leanhelix.SenderSignature, block leanhelix.Block) leanhelix.ViewChangeMessage {
	panic("this is Orbs code")
}

func (m *MockMessageFactory) CreateNewViewMessage(preprepareMessage leanhelix.PreprepareMessage, nvHeader leanhelix.NewViewHeader, sender leanhelix.SenderSignature) leanhelix.NewViewMessage {
	panic("this is Orbs code")
}

// Auxiliary methods

func (m *MockMessageFactory) CreateSenderSignature(sender []byte, signature []byte) leanhelix.SenderSignature {
	panic("this is Orbs code")
}

func (m *MockMessageFactory) CreateBlockRef(messageType int, blockHeight int, view int, blockHash []byte) leanhelix.BlockRef {
	panic("this is Orbs code")
}
func (m *MockMessageFactory) CreateNewViewHeader(messageType int, blockHeight int, view int, confirmations []leanhelix.ViewChangeConfirmation) leanhelix.NewViewHeader {
	panic("this is Orbs code")
}
func (m *MockMessageFactory) CreateViewChangeConfirmation(vcHeader leanhelix.ViewChangeHeader, sender leanhelix.SenderSignature) leanhelix.ViewChangeConfirmation {
	panic("this is Orbs code")
}
func (m *MockMessageFactory) CreateViewChangeHeader(blockHeight int, view int, proof leanhelix.PreparedProof) leanhelix.ViewChangeHeader {
	panic("this is Orbs code")
}
func (m *MockMessageFactory) CreatePreparedProof(ppBlockRef leanhelix.BlockRef, pBlockRef leanhelix.BlockRef, ppSender leanhelix.SenderSignature, pSenders []leanhelix.SenderSignature) leanhelix.PreparedProof {
	panic("this is Orbs code")
}

func TestKeyManagerImpl(t *testing.T) {

}

func CreatePPM() leanhelix.PreprepareMessage {
	messageFactory := &MockMessageFactory{}
	keyManager := &MockKeyManager{}
	messageType := 0
	height := 1
	view := 2
	blockHash := []byte{10, 20, 30}
	blockRef := messageFactory.CreateBlockRef(messageType, height, view, blockHash)
	block := builders.CreateBlock(builders.GenesisBlock)
	sig := keyManager.Sign(blockRef.Serialize())
	senderSignature := messageFactory.CreateSenderSignature([]byte("MyPK"), sig)
	ppm := messageFactory.CreatePreprepareMessage(blockRef, senderSignature, block)

	return ppm

}

func CreatePM() leanhelix.PrepareMessage {
	messageFactory := &MockMessageFactory{}
	keyManager := &MockKeyManager{}
	messageType := 1
	height := 1
	view := 2
	blockHash := []byte{10, 20, 30}
	blockRef := messageFactory.CreateBlockRef(messageType, height, view, blockHash)
	sig := keyManager.Sign(blockRef.Serialize())
	senderSignature := messageFactory.CreateSenderSignature([]byte("MyPK"), sig)
	pm := messageFactory.CreatePrepareMessage(blockRef, senderSignature)

	return pm

}

func TestCreateAndSignPPMessage(t *testing.T) {

	ppm := CreatePPM()
	fmt.Println(ppm)

}

func TestCreateAndSignNVMessage(t *testing.T) {

	messageFactory := &MockMessageFactory{}
	keyManager := &MockKeyManager{}
	messageType := 0
	height := 1
	view := 2
	//blockHash := []byte{10, 20, 30}

	ppm := CreatePPM()
	pm := CreatePM()
	proof := messageFactory.CreatePreparedProof(ppm.SignedHeader(), pm.SignedHeader(), ppm.Sender(), []leanhelix.SenderSignature{pm.Sender()})
	vcHeader := messageFactory.CreateViewChangeHeader(height, view, proof)
	vcSig := keyManager.Sign(vcHeader.Serialize())
	vcSenderSignature := messageFactory.CreateSenderSignature([]byte("MyPK"), vcSig)
	confirmation := messageFactory.CreateViewChangeConfirmation(vcHeader, vcSenderSignature)
	nvHeader := messageFactory.CreateNewViewHeader(messageType, height, view, []leanhelix.ViewChangeConfirmation{confirmation})
	nvSig := keyManager.Sign(nvHeader.Serialize())
	nvSenderSignature := messageFactory.CreateSenderSignature([]byte("MyPK"), nvSig)
	nvm := messageFactory.CreateNewViewMessage(ppm, nvHeader, nvSenderSignature)

	fmt.Println(nvm)

}

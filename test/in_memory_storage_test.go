package test

import (
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestStorePreprepare(t *testing.T) {
	var storage lh.Storage = lh.NewInMemoryStorage()
	blockHeight := BlockHeight(math.Floor(rand.Float64() * 1000))
	view := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	keyManager1 := builders.NewMockKeyManager(senderId1)
	keyManager2 := builders.NewMockKeyManager(senderId2)
	block := builders.CreateBlock(builders.GenesisBlock)

	mf1 := lh.NewMessageFactory(keyManager1)
	preprepareMessage1 := mf1.CreatePreprepareMessage(blockHeight, view, block)

	mf2 := lh.NewMessageFactory(keyManager2)
	preprepareMessage2 := mf2.CreatePreprepareMessage(blockHeight, view, block)

	storage.StorePreprepare(preprepareMessage1)
	storage.StorePreprepare(preprepareMessage2)

	actualPreprepareMessage, _ := storage.GetPreprepareMessage(blockHeight, view)
	actualPreprepareBlock, _ := storage.GetPreprepareBlock(blockHeight, view)

	require.Equal(t, actualPreprepareMessage, preprepareMessage1, "stored preprepare message should match the fetched preprepare message")
	require.Equal(t, actualPreprepareBlock, block, "stored preprepare block should match the fetched preprepare block")
}

func TestStorePrepare(t *testing.T) {
	var storage lh.Storage = lh.NewInMemoryStorage()
	blockHeight1 := BlockHeight(math.Floor(rand.Float64() * 1000))
	blockHeight2 := BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := View(math.Floor(rand.Float64() * 1000))
	view2 := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	keyManager1 := builders.NewMockKeyManager(senderId1)
	keyManager2 := builders.NewMockKeyManager(senderId2)
	keyManager3 := builders.NewMockKeyManager(senderId3)
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(builders.GenesisBlock)

	mf1 := lh.NewMessageFactory(keyManager1)
	mf2 := lh.NewMessageFactory(keyManager2)
	mf3 := lh.NewMessageFactory(keyManager3)

	message1 := mf1.CreatePrepareMessage(blockHeight1, view1, block1.BlockHash())
	message2 := mf2.CreatePrepareMessage(blockHeight1, view1, block1.BlockHash())
	message3 := mf3.CreatePrepareMessage(blockHeight1, view1, block1.BlockHash())
	message4 := mf1.CreatePrepareMessage(blockHeight2, view1, block1.BlockHash())
	message5 := mf1.CreatePrepareMessage(blockHeight1, view2, block1.BlockHash())
	message6 := mf1.CreatePrepareMessage(blockHeight1, view1, block2.BlockHash())

	storage.StorePrepare(message1)
	storage.StorePrepare(message2)
	storage.StorePrepare(message3)
	storage.StorePrepare(message4)
	storage.StorePrepare(message5)
	storage.StorePrepare(message6)

	actualPrepareMessages, _ := storage.GetPrepareMessages(blockHeight1, view1, block1.BlockHash())
	expectedMessages := []*lh.PrepareMessage{message1, message2, message3}
	require.ElementsMatch(t, actualPrepareMessages, expectedMessages, "stored prepare messages should match the fetched prepare messages")

	actualPrepareSendersPks := storage.GetPrepareSendersPKs(blockHeight1, view1, block1.BlockHash())
	expectedPks := []Ed25519PublicKey{senderId1, senderId2, senderId3}
	require.ElementsMatch(t, actualPrepareSendersPks, expectedPks, "stored prepare messages senders should match the fetched prepare messages senders")
}

func TestStoreCommit(t *testing.T) {
	var storage lh.Storage = lh.NewInMemoryStorage()
	blockHeight1 := BlockHeight(math.Floor(rand.Float64() * 1000))
	blockHeight2 := BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := View(math.Floor(rand.Float64() * 1000))
	view2 := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	keyManager1 := builders.NewMockKeyManager(senderId1)
	keyManager2 := builders.NewMockKeyManager(senderId2)
	keyManager3 := builders.NewMockKeyManager(senderId3)
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(builders.GenesisBlock)

	mf1 := lh.NewMessageFactory(keyManager1)
	mf2 := lh.NewMessageFactory(keyManager2)
	mf3 := lh.NewMessageFactory(keyManager3)

	message1 := mf1.CreateCommitMessage(blockHeight1, view1, block1.BlockHash())
	message2 := mf2.CreateCommitMessage(blockHeight1, view1, block1.BlockHash())
	message3 := mf3.CreateCommitMessage(blockHeight1, view1, block1.BlockHash())
	message4 := mf1.CreateCommitMessage(blockHeight2, view1, block1.BlockHash())
	message5 := mf1.CreateCommitMessage(blockHeight1, view2, block1.BlockHash())
	message6 := mf1.CreateCommitMessage(blockHeight1, view1, block2.BlockHash())

	storage.StoreCommit(message1)
	storage.StoreCommit(message2)
	storage.StoreCommit(message3)
	storage.StoreCommit(message4)
	storage.StoreCommit(message5)
	storage.StoreCommit(message6)

	actualCommitMessages, _ := storage.GetCommitMessages(blockHeight1, view1, block1.BlockHash())
	expectedMessages := []*lh.CommitMessage{message1, message2, message3}
	require.ElementsMatch(t, actualCommitMessages, expectedMessages, "stored commit messages should match the fetched commit messages")

	actualCommitSendersPks := storage.GetCommitSendersPKs(blockHeight1, view1, block1.BlockHash())
	expectedPks := []Ed25519PublicKey{senderId1, senderId2, senderId3}
	require.ElementsMatch(t, actualCommitSendersPks, expectedPks, "stored commit messages senders should match the fetched commit messages senders")
}

func TestStoreViewChange(t *testing.T) {
	var storage lh.Storage = lh.NewInMemoryStorage()
	blockHeight1 := BlockHeight(math.Floor(rand.Float64() * 1000))
	blockHeight2 := BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := View(math.Floor(rand.Float64() * 1000))
	view2 := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	keyManager1 := builders.NewMockKeyManager(senderId1)
	keyManager2 := builders.NewMockKeyManager(senderId2)
	keyManager3 := builders.NewMockKeyManager(senderId3)

	mf1 := lh.NewMessageFactory(keyManager1)
	mf2 := lh.NewMessageFactory(keyManager2)
	mf3 := lh.NewMessageFactory(keyManager3)

	message1 := mf1.CreateViewChangeMessage(blockHeight1, view1, nil)
	message2 := mf2.CreateViewChangeMessage(blockHeight1, view1, nil)
	message3 := mf3.CreateViewChangeMessage(blockHeight1, view1, nil)
	message4 := mf1.CreateViewChangeMessage(blockHeight2, view1, nil)
	message5 := mf1.CreateViewChangeMessage(blockHeight1, view2, nil)

	storage.StoreViewChange(message1)
	storage.StoreViewChange(message2)
	storage.StoreViewChange(message3)
	storage.StoreViewChange(message4)
	storage.StoreViewChange(message5)

	actualViewChangeMessages := storage.GetViewChangeMessages(blockHeight1, view1)
	expectedMessages := []*lh.ViewChangeMessage{message1, message2, message3}
	require.ElementsMatch(t, actualViewChangeMessages, expectedMessages, "stored view-change messages should match the fetched view-change messages")
}

func TestDuplicatePreprepare(t *testing.T) {
	var storage lh.Storage = lh.NewInMemoryStorage()
	block := builders.CreateBlock(builders.GenesisBlock)
	keyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK"))
	mf := lh.NewMessageFactory(keyManager)
	ppm := mf.CreatePreprepareMessage(1, 1, block)

	firstTime := storage.StorePreprepare(ppm)
	require.True(t, firstTime, "StorePreprepare() returns true if storing a new value ")

	secondTime := storage.StorePreprepare(ppm)
	require.False(t, secondTime, "StorePreprepare() returns false if trying to store a value that already exists")
}

func TestDuplicatePrepare(t *testing.T) {
	var storage lh.Storage = lh.NewInMemoryStorage()
	blockHeight := BlockHeight(math.Floor(rand.Float64() * 1000))
	view := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(senderId1)
	sender2KeyManager := builders.NewMockKeyManager(senderId2)
	block := builders.CreateBlock(builders.GenesisBlock)
	sender1MsgFactory := lh.NewMessageFactory(sender1KeyManager)
	sender2MsgFactory := lh.NewMessageFactory(sender2KeyManager)
	p1 := sender1MsgFactory.CreatePrepareMessage(blockHeight, view, block.BlockHash())
	p2 := sender2MsgFactory.CreatePrepareMessage(blockHeight, view, block.BlockHash())

	firstTime := storage.StorePrepare(p1)
	require.True(t, firstTime, "StorePrepare() returns true if storing a new value (1 of 2)")

	secondTime := storage.StorePrepare(p2)
	require.True(t, secondTime, "StorePrepare() returns true if storing a new value (2 of 2)")

	thirdTime := storage.StorePrepare(p2)
	require.False(t, thirdTime, "StorePrepare() returns false if trying to store a value that already exists")
}

func TestDuplicateCommit(t *testing.T) {
	var storage lh.Storage = lh.NewInMemoryStorage()
	blockHeight := BlockHeight(math.Floor(rand.Float64() * 1000))
	view := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(senderId1)
	sender2KeyManager := builders.NewMockKeyManager(senderId2)
	block := builders.CreateBlock(builders.GenesisBlock)
	sender1MsgFactory := lh.NewMessageFactory(sender1KeyManager)
	sender2MsgFactory := lh.NewMessageFactory(sender2KeyManager)

	c1 := sender1MsgFactory.CreateCommitMessage(blockHeight, view, block.BlockHash())
	c2 := sender2MsgFactory.CreateCommitMessage(blockHeight, view, block.BlockHash())

	firstTime := storage.StoreCommit(c1)
	require.True(t, firstTime, "StoreCommit() returns true if storing a new value (1 of 2)")

	secondTime := storage.StoreCommit(c2)
	require.True(t, secondTime, "StoreCommit() returns true if storing a new value (2 of 2)")

	thirdTime := storage.StoreCommit(c2)
	require.False(t, thirdTime, "StoreCommit() returns false if trying to store a value that already exists")

}

func TestDuplicateViewChange(t *testing.T) {
	var storage lh.Storage = lh.NewInMemoryStorage()
	blockHeight := BlockHeight(math.Floor(rand.Float64() * 1000))
	view := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(senderId1)
	sender2KeyManager := builders.NewMockKeyManager(senderId2)
	sender1MsgFactory := lh.NewMessageFactory(sender1KeyManager)
	sender2MsgFactory := lh.NewMessageFactory(sender2KeyManager)
	vc1 := sender1MsgFactory.CreateViewChangeMessage(blockHeight, view, nil)
	vc2 := sender2MsgFactory.CreateViewChangeMessage(blockHeight, view, nil)

	firstTime := storage.StoreViewChange(vc1)
	require.True(t, firstTime, "StoreViewChange() returns true if storing a new value (1 of 2)")

	secondTime := storage.StoreViewChange(vc2)
	require.True(t, secondTime, "StoreViewChange() returns true if storing a new value (2 of 2)")

	thirdTime := storage.StoreViewChange(vc2)
	require.False(t, thirdTime, "StoreViewChange() returns false if trying to store a value that already exists")

}

func TestClearBlockHeightLogs(t *testing.T) {
	var storage lh.Storage = lh.NewInMemoryStorage()
	blockHeight := BlockHeight(math.Floor(rand.Float64() * 1000))
	view := View(math.Floor(rand.Float64() * 1000))
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := builders.CalculateBlockHash(block)
	keyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK"))

	msgFactory := lh.NewMessageFactory(keyManager)
	ppMsg := msgFactory.CreatePreprepareMessage(blockHeight, view, block)
	pMsg := msgFactory.CreatePrepareMessage(blockHeight, view, blockHash)
	cMsg := msgFactory.CreateCommitMessage(blockHeight, view, blockHash)
	vcMsg := msgFactory.CreateViewChangeMessage(blockHeight, view, nil)

	storage.StorePreprepare(ppMsg)
	storage.StorePrepare(pMsg)
	storage.StoreCommit(cMsg)
	storage.StoreViewChange(vcMsg)

	actualPP, _ := storage.GetPreprepareMessage(blockHeight, view)
	actualP, _ := storage.GetPrepareMessages(blockHeight, view, blockHash)
	actualC, _ := storage.GetCommitMessages(blockHeight, view, blockHash)
	actualVC := storage.GetViewChangeMessages(blockHeight, view)
	require.Equal(t, actualPP, ppMsg, "stored preprepare message should match the fetched preprepare message")
	require.Equal(t, 1, len(actualP), "Length of GetPrepareMessages() result array should be 1")
	require.Equal(t, 1, len(actualC), "Length of GetCommitSendersPKs() result array should be 1")
	require.Equal(t, 1, len(actualVC), "Length of GetViewChangeMessages() result array should be 1")

	storage.ClearBlockHeightLogs(blockHeight)

	actualPP, _ = storage.GetPreprepareMessage(blockHeight, view)
	actualP, _ = storage.GetPrepareMessages(blockHeight, view, blockHash)
	actualC, _ = storage.GetCommitMessages(blockHeight, view, blockHash)
	actualVC = storage.GetViewChangeMessages(blockHeight, view)

	require.Nil(t, actualPP, "GetPreprepareMessage() should return nil after ClearBlockHeightLogs()")
	require.Equal(t, 0, len(actualP), "Length of GetPrepareMessages() result array should be 0")
	require.Equal(t, 0, len(actualC), "Length of GetCommitSendersPKs() result array should be 0")
	require.Equal(t, 0, len(actualVC), "Length of GetViewChangeMessages() result array should be 0")
}

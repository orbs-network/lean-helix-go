package test

import (
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func TestStorePreprepare(t *testing.T) {
	var storage leanhelix.Storage = leanhelix.NewInMemoryStorage()
	blockHeight := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := primitives.View(math.Floor(rand.Float64() * 1000000))
	senderId1 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId2 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	keyManager1 := builders.NewMockKeyManager(senderId1)
	keyManager2 := builders.NewMockKeyManager(senderId2)
	block := builders.CreateBlock(builders.GenesisBlock)

	preprepareMessage1 := builders.APreprepareMessage(keyManager1, senderId1, blockHeight, view, block)
	preprepareMessage2 := builders.APreprepareMessage(keyManager2, senderId2, blockHeight, view, block)

	storage.StorePreprepare(preprepareMessage1)
	storage.StorePreprepare(preprepareMessage2)

	actualPreprepareMessage, _ := storage.GetPreprepareMessage(blockHeight, view)
	actualPreprepareBlock, _ := storage.GetPreprepareBlock(blockHeight, view)

	require.Equal(t, actualPreprepareMessage, preprepareMessage1, "stored preprepare message should match the fetched preprepare message")
	require.Equal(t, actualPreprepareBlock, block, "stored preprepare block should match the fetched preprepare block")
}

func TestStorePrepare(t *testing.T) {
	var storage leanhelix.Storage = leanhelix.NewInMemoryStorage()
	blockHeight1 := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	blockHeight2 := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view1 := primitives.View(math.Floor(rand.Float64() * 1000000))
	view2 := primitives.View(math.Floor(rand.Float64() * 1000000))
	senderId1 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId2 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId3 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	keyManager1 := builders.NewMockKeyManager(senderId1)
	keyManager2 := builders.NewMockKeyManager(senderId2)
	keyManager3 := builders.NewMockKeyManager(senderId3)
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(builders.GenesisBlock)
	block1Hash := builders.CalculateBlockHash(block1)

	message1 := builders.APrepareMessage(keyManager1, senderId1, blockHeight1, view1, block1)
	message2 := builders.APrepareMessage(keyManager2, senderId2, blockHeight1, view1, block1)
	message3 := builders.APrepareMessage(keyManager3, senderId3, blockHeight1, view1, block1)
	message4 := builders.APrepareMessage(keyManager1, senderId1, blockHeight2, view1, block1)
	message5 := builders.APrepareMessage(keyManager1, senderId1, blockHeight1, view2, block1)
	message6 := builders.APrepareMessage(keyManager1, senderId1, blockHeight1, view1, block2)

	storage.StorePrepare(message1)
	storage.StorePrepare(message2)
	storage.StorePrepare(message3)
	storage.StorePrepare(message4)
	storage.StorePrepare(message5)
	storage.StorePrepare(message6)

	actualPrepareMessages, _ := storage.GetPrepareMessages(blockHeight1, view1, block1Hash)
	expectedMessages := []*leanhelix.PrepareMessage{message1, message2, message3}
	require.ElementsMatch(t, actualPrepareMessages, expectedMessages, "stored prepare messages should match the fetched prepare messages")

	actualPrepareSendersIds := storage.GetPrepareSendersIds(blockHeight1, view1, block1Hash)
	expectedIds := []primitives.MemberId{senderId1, senderId2, senderId3}
	require.ElementsMatch(t, actualPrepareSendersIds, expectedIds, "stored prepare messages senders should match the fetched prepare messages senders")
}

func TestStoreCommit(t *testing.T) {
	var storage leanhelix.Storage = leanhelix.NewInMemoryStorage()
	blockHeight1 := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	blockHeight2 := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view1 := primitives.View(math.Floor(rand.Float64() * 1000000))
	view2 := primitives.View(math.Floor(rand.Float64() * 1000000))
	senderId1 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId2 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId3 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	keyManager1 := builders.NewMockKeyManager(senderId1)
	keyManager2 := builders.NewMockKeyManager(senderId2)
	keyManager3 := builders.NewMockKeyManager(senderId3)
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(builders.GenesisBlock)
	block1Hash := builders.CalculateBlockHash(block1)

	message1 := builders.ACommitMessage(keyManager1, senderId1, blockHeight1, view1, block1)
	message2 := builders.ACommitMessage(keyManager2, senderId2, blockHeight1, view1, block1)
	message3 := builders.ACommitMessage(keyManager3, senderId3, blockHeight1, view1, block1)
	message4 := builders.ACommitMessage(keyManager1, senderId1, blockHeight2, view1, block1)
	message5 := builders.ACommitMessage(keyManager1, senderId1, blockHeight1, view2, block1)
	message6 := builders.ACommitMessage(keyManager1, senderId1, blockHeight1, view1, block2)

	storage.StoreCommit(message1)
	storage.StoreCommit(message2)
	storage.StoreCommit(message3)
	storage.StoreCommit(message4)
	storage.StoreCommit(message5)
	storage.StoreCommit(message6)

	actualCommitMessages, _ := storage.GetCommitMessages(blockHeight1, view1, block1Hash)
	expectedMessages := []*leanhelix.CommitMessage{message1, message2, message3}
	require.ElementsMatch(t, actualCommitMessages, expectedMessages, "stored commit messages should match the fetched commit messages")

	actualCommitSendersIds := storage.GetCommitSendersIds(blockHeight1, view1, block1Hash)
	expectedIds := []primitives.MemberId{senderId1, senderId2, senderId3}
	require.ElementsMatch(t, actualCommitSendersIds, expectedIds, "stored commit messages senders should match the fetched commit messages senders")
}

func TestStoreViewChange(t *testing.T) {
	var storage leanhelix.Storage = leanhelix.NewInMemoryStorage()
	blockHeight1 := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	blockHeight2 := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view1 := primitives.View(math.Floor(rand.Float64() * 1000000))
	view2 := primitives.View(math.Floor(rand.Float64() * 1000000))
	senderId1 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId2 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId3 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	keyManager1 := builders.NewMockKeyManager(senderId1)
	keyManager2 := builders.NewMockKeyManager(senderId2)
	keyManager3 := builders.NewMockKeyManager(senderId3)

	message1 := builders.AViewChangeMessage(keyManager1, senderId1, blockHeight1, view1, nil)
	message2 := builders.AViewChangeMessage(keyManager2, senderId2, blockHeight1, view1, nil)
	message3 := builders.AViewChangeMessage(keyManager3, senderId3, blockHeight1, view1, nil)
	message4 := builders.AViewChangeMessage(keyManager1, senderId1, blockHeight2, view1, nil)
	message5 := builders.AViewChangeMessage(keyManager1, senderId1, blockHeight1, view2, nil)

	storage.StoreViewChange(message1)
	storage.StoreViewChange(message2)
	storage.StoreViewChange(message3)
	storage.StoreViewChange(message4)
	storage.StoreViewChange(message5)

	actualViewChangeMessages, _ := storage.GetViewChangeMessages(blockHeight1, view1)
	expectedMessages := []*leanhelix.ViewChangeMessage{message1, message2, message3}
	require.ElementsMatch(t, actualViewChangeMessages, expectedMessages, "stored view-change messages should match the fetched view-change messages")
}

func TestLatestPreprepare(t *testing.T) {
	var storage leanhelix.Storage = leanhelix.NewInMemoryStorage()
	blockHeight := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	senderId1 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId2 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	keyManager1 := builders.NewMockKeyManager(senderId1)
	keyManager2 := builders.NewMockKeyManager(senderId2)
	block := builders.CreateBlock(builders.GenesisBlock)

	preprepareMessageOnView3 := builders.APreprepareMessage(keyManager1, senderId1, blockHeight, 3, block)
	preprepareMessageOnView2 := builders.APreprepareMessage(keyManager2, senderId2, blockHeight, 2, block)

	storage.StorePreprepare(preprepareMessageOnView3)
	storage.StorePreprepare(preprepareMessageOnView2)

	actualLatestPreprepareMessage, _ := storage.GetLatestPreprepare(blockHeight)

	require.Equal(t, actualLatestPreprepareMessage, preprepareMessageOnView3, "fetching preprepare should return the latest preprepare")
}

func TestDuplicatePreprepare(t *testing.T) {
	var storage leanhelix.Storage = leanhelix.NewInMemoryStorage()
	block := builders.CreateBlock(builders.GenesisBlock)
	memberId := primitives.MemberId("Member Id")
	keyManager := builders.NewMockKeyManager(memberId)
	ppm := builders.APreprepareMessage(keyManager, memberId, 1, 1, block)

	firstTime := storage.StorePreprepare(ppm)
	require.True(t, firstTime, "StorePreprepare() returns true if storing a new value ")

	secondTime := storage.StorePreprepare(ppm)
	require.False(t, secondTime, "StorePreprepare() returns false if trying to store a value that already exists")
}

func TestDuplicatePrepare(t *testing.T) {
	var storage leanhelix.Storage = leanhelix.NewInMemoryStorage()
	blockHeight := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := primitives.View(math.Floor(rand.Float64() * 1000000))
	senderId1 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId2 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	sender1KeyManager := builders.NewMockKeyManager(senderId1)
	sender2KeyManager := builders.NewMockKeyManager(senderId2)
	block := builders.CreateBlock(builders.GenesisBlock)
	p1 := builders.APrepareMessage(sender1KeyManager, senderId1, blockHeight, view, block)
	p2 := builders.APrepareMessage(sender2KeyManager, senderId2, blockHeight, view, block)

	firstTime := storage.StorePrepare(p1)
	require.True(t, firstTime, "StorePrepare() returns true if storing a new value (1 of 2)")

	secondTime := storage.StorePrepare(p2)
	require.True(t, secondTime, "StorePrepare() returns true if storing a new value (2 of 2)")

	thirdTime := storage.StorePrepare(p2)
	require.False(t, thirdTime, "StorePrepare() returns false if trying to store a value that already exists")
}

func TestDuplicateCommit(t *testing.T) {
	var storage leanhelix.Storage = leanhelix.NewInMemoryStorage()
	blockHeight := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := primitives.View(math.Floor(rand.Float64() * 1000000))
	senderId1 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId2 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	sender1KeyManager := builders.NewMockKeyManager(senderId1)
	sender2KeyManager := builders.NewMockKeyManager(senderId2)
	block := builders.CreateBlock(builders.GenesisBlock)

	c1 := builders.ACommitMessage(sender1KeyManager, senderId1, blockHeight, view, block)
	c2 := builders.ACommitMessage(sender2KeyManager, senderId2, blockHeight, view, block)

	firstTime := storage.StoreCommit(c1)
	require.True(t, firstTime, "StoreCommit() returns true if storing a new value (1 of 2)")

	secondTime := storage.StoreCommit(c2)
	require.True(t, secondTime, "StoreCommit() returns true if storing a new value (2 of 2)")

	thirdTime := storage.StoreCommit(c2)
	require.False(t, thirdTime, "StoreCommit() returns false if trying to store a value that already exists")

}

func TestDuplicateViewChange(t *testing.T) {
	var storage leanhelix.Storage = leanhelix.NewInMemoryStorage()
	blockHeight := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := primitives.View(math.Floor(rand.Float64() * 1000000))
	senderId1 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	senderId2 := primitives.MemberId(strconv.Itoa(int(math.Floor(rand.Float64() * 1000000))))
	sender1KeyManager := builders.NewMockKeyManager(senderId1)
	sender2KeyManager := builders.NewMockKeyManager(senderId2)
	vc1 := builders.AViewChangeMessage(sender1KeyManager, senderId1, blockHeight, view, nil)
	vc2 := builders.AViewChangeMessage(sender2KeyManager, senderId2, blockHeight, view, nil)

	firstTime := storage.StoreViewChange(vc1)
	require.True(t, firstTime, "StoreViewChange() returns true if storing a new value (1 of 2)")

	secondTime := storage.StoreViewChange(vc2)
	require.True(t, secondTime, "StoreViewChange() returns true if storing a new value (2 of 2)")

	thirdTime := storage.StoreViewChange(vc2)
	require.False(t, thirdTime, "StoreViewChange() returns false if trying to store a value that already exists")

}

func TestClearBlockHeightLogs(t *testing.T) {
	var storage leanhelix.Storage = leanhelix.NewInMemoryStorage()
	blockHeight := primitives.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := primitives.View(math.Floor(rand.Float64() * 1000000))
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := builders.CalculateBlockHash(block)
	memberId := primitives.MemberId("Member Id")
	keyManager := builders.NewMockKeyManager(memberId)

	ppMsg := builders.APreprepareMessage(keyManager, memberId, blockHeight, view, block)
	pMsg := builders.APrepareMessage(keyManager, memberId, blockHeight, view, block)
	cMsg := builders.ACommitMessage(keyManager, memberId, blockHeight, view, block)
	vcMsg := builders.AViewChangeMessage(keyManager, memberId, blockHeight, view, nil)

	storage.StorePreprepare(ppMsg)
	storage.StorePrepare(pMsg)
	storage.StoreCommit(cMsg)
	storage.StoreViewChange(vcMsg)

	actualPP, _ := storage.GetPreprepareMessage(blockHeight, view)
	actualP, _ := storage.GetPrepareMessages(blockHeight, view, blockHash)
	actualC, _ := storage.GetCommitMessages(blockHeight, view, blockHash)
	actualVC, _ := storage.GetViewChangeMessages(blockHeight, view)
	require.Equal(t, actualPP, ppMsg, "stored preprepare message should match the fetched preprepare message")
	require.Equal(t, 1, len(actualP), "Length of GetPrepareMessages() result array should be 1")
	require.Equal(t, 1, len(actualC), "Length of GetCommitSendersIds() result array should be 1")
	require.Equal(t, 1, len(actualVC), "Length of GetViewChangeMessages() result array should be 1")

	storage.ClearBlockHeightLogs(blockHeight)

	actualPP, _ = storage.GetPreprepareMessage(blockHeight, view)
	actualP, _ = storage.GetPrepareMessages(blockHeight, view, blockHash)
	actualC, _ = storage.GetCommitMessages(blockHeight, view, blockHash)
	actualVC, _ = storage.GetViewChangeMessages(blockHeight, view)

	require.Nil(t, actualPP, "GetPreprepareMessage() should return nil after ClearBlockHeightLogs()")
	require.Equal(t, 0, len(actualP), "Length of GetPrepareMessages() result array should be 0")
	require.Equal(t, 0, len(actualC), "Length of GetCommitSendersIds() result array should be 0")
	require.Equal(t, 0, len(actualVC), "Length of GetViewChangeMessages() result array should be 0")
}

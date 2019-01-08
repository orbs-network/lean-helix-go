package test

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/services/storage"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/mocks"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"testing"
)

func TestStorePreprepare(t *testing.T) {
	var s interfaces.Storage = storage.NewInMemoryStorage()
	instanceId := primitives.InstanceId(rand.Uint64())
	blockHeight := primitives.BlockHeight(rand.Uint64())
	view := primitives.View(rand.Uint64())
	senderId1 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId2 := primitives.MemberId(strconv.Itoa(rand.Int()))
	keyManager1 := mocks.NewMockKeyManager(senderId1)
	keyManager2 := mocks.NewMockKeyManager(senderId2)
	block := mocks.ABlock(interfaces.GenesisBlock)

	preprepareMessage1 := builders.APreprepareMessage(instanceId, keyManager1, senderId1, blockHeight, view, block)
	preprepareMessage2 := builders.APreprepareMessage(instanceId, keyManager2, senderId2, blockHeight, view, block)

	s.StorePreprepare(preprepareMessage1)
	s.StorePreprepare(preprepareMessage2)

	actualPreprepareMessage, _ := s.GetPreprepareMessage(blockHeight, view)
	actualPreprepareBlock, _ := s.GetPreprepareBlock(blockHeight, view)

	require.Equal(t, actualPreprepareMessage, preprepareMessage1, "stored preprepare message should match the fetched preprepare message")
	require.Equal(t, actualPreprepareBlock, block, "stored preprepare block should match the fetched preprepare block")
}

func TestStorePrepare(t *testing.T) {
	var s interfaces.Storage = storage.NewInMemoryStorage()
	instanceId := primitives.InstanceId(rand.Uint64())
	blockHeight1 := primitives.BlockHeight(rand.Uint64())
	blockHeight2 := primitives.BlockHeight(rand.Uint64())
	view1 := primitives.View(rand.Uint64())
	view2 := primitives.View(rand.Uint64())
	senderId1 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId2 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId3 := primitives.MemberId(strconv.Itoa(rand.Int()))
	keyManager1 := mocks.NewMockKeyManager(senderId1)
	keyManager2 := mocks.NewMockKeyManager(senderId2)
	keyManager3 := mocks.NewMockKeyManager(senderId3)
	block1 := mocks.ABlock(interfaces.GenesisBlock)
	block2 := mocks.ABlock(interfaces.GenesisBlock)
	block1Hash := mocks.CalculateBlockHash(block1)

	message1 := builders.APrepareMessage(instanceId, keyManager1, senderId1, blockHeight1, view1, block1)
	message2 := builders.APrepareMessage(instanceId, keyManager2, senderId2, blockHeight1, view1, block1)
	message3 := builders.APrepareMessage(instanceId, keyManager3, senderId3, blockHeight1, view1, block1)
	message4 := builders.APrepareMessage(instanceId, keyManager1, senderId1, blockHeight2, view1, block1)
	message5 := builders.APrepareMessage(instanceId, keyManager1, senderId1, blockHeight1, view2, block1)
	message6 := builders.APrepareMessage(instanceId, keyManager1, senderId1, blockHeight1, view1, block2)

	s.StorePrepare(message1)
	s.StorePrepare(message2)
	s.StorePrepare(message3)
	s.StorePrepare(message4)
	s.StorePrepare(message5)
	s.StorePrepare(message6)

	actualPrepareMessages, _ := s.GetPrepareMessages(blockHeight1, view1, block1Hash)
	expectedMessages := []*interfaces.PrepareMessage{message1, message2, message3}
	require.ElementsMatch(t, actualPrepareMessages, expectedMessages, "stored prepare messages should match the fetched prepare messages")

	actualPrepareSendersIds := s.GetPrepareSendersIds(blockHeight1, view1, block1Hash)
	expectedIds := []primitives.MemberId{senderId1, senderId2, senderId3}
	require.ElementsMatch(t, actualPrepareSendersIds, expectedIds, "stored prepare messages senders should match the fetched prepare messages senders")
}

func TestStoreCommit(t *testing.T) {
	var s interfaces.Storage = storage.NewInMemoryStorage()
	instanceId := primitives.InstanceId(rand.Uint64())
	blockHeight1 := primitives.BlockHeight(rand.Uint64())
	blockHeight2 := primitives.BlockHeight(rand.Uint64())
	view1 := primitives.View(rand.Uint64())
	view2 := primitives.View(rand.Uint64())
	senderId1 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId2 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId3 := primitives.MemberId(strconv.Itoa(rand.Int()))
	keyManager1 := mocks.NewMockKeyManager(senderId1)
	keyManager2 := mocks.NewMockKeyManager(senderId2)
	keyManager3 := mocks.NewMockKeyManager(senderId3)
	block1 := mocks.ABlock(interfaces.GenesisBlock)
	block2 := mocks.ABlock(interfaces.GenesisBlock)
	block1Hash := mocks.CalculateBlockHash(block1)

	message1 := builders.ACommitMessage(instanceId, keyManager1, senderId1, blockHeight1, view1, block1, 0)
	message2 := builders.ACommitMessage(instanceId, keyManager2, senderId2, blockHeight1, view1, block1, 0)
	message3 := builders.ACommitMessage(instanceId, keyManager3, senderId3, blockHeight1, view1, block1, 0)
	message4 := builders.ACommitMessage(instanceId, keyManager1, senderId1, blockHeight2, view1, block1, 0)
	message5 := builders.ACommitMessage(instanceId, keyManager1, senderId1, blockHeight1, view2, block1, 0)
	message6 := builders.ACommitMessage(instanceId, keyManager1, senderId1, blockHeight1, view1, block2, 0)

	s.StoreCommit(message1)
	s.StoreCommit(message2)
	s.StoreCommit(message3)
	s.StoreCommit(message4)
	s.StoreCommit(message5)
	s.StoreCommit(message6)

	actualCommitMessages, _ := s.GetCommitMessages(blockHeight1, view1, block1Hash)
	expectedMessages := []*interfaces.CommitMessage{message1, message2, message3}
	require.ElementsMatch(t, actualCommitMessages, expectedMessages, "stored commit messages should match the fetched commit messages")
}

func TestStoreViewChange(t *testing.T) {
	var s interfaces.Storage = storage.NewInMemoryStorage()
	instanceId := primitives.InstanceId(rand.Uint64())
	blockHeight1 := primitives.BlockHeight(rand.Uint64())
	blockHeight2 := primitives.BlockHeight(rand.Uint64())
	view1 := primitives.View(rand.Uint64())
	view2 := primitives.View(rand.Uint64())
	senderId1 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId2 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId3 := primitives.MemberId(strconv.Itoa(rand.Int()))
	keyManager1 := mocks.NewMockKeyManager(senderId1)
	keyManager2 := mocks.NewMockKeyManager(senderId2)
	keyManager3 := mocks.NewMockKeyManager(senderId3)

	message1 := builders.AViewChangeMessage(instanceId, keyManager1, senderId1, blockHeight1, view1, nil)
	message2 := builders.AViewChangeMessage(instanceId, keyManager2, senderId2, blockHeight1, view1, nil)
	message3 := builders.AViewChangeMessage(instanceId, keyManager3, senderId3, blockHeight1, view1, nil)
	message4 := builders.AViewChangeMessage(instanceId, keyManager1, senderId1, blockHeight2, view1, nil)
	message5 := builders.AViewChangeMessage(instanceId, keyManager1, senderId1, blockHeight1, view2, nil)

	s.StoreViewChange(message1)
	s.StoreViewChange(message2)
	s.StoreViewChange(message3)
	s.StoreViewChange(message4)
	s.StoreViewChange(message5)

	actualViewChangeMessages, _ := s.GetViewChangeMessages(blockHeight1, view1)
	expectedMessages := []*interfaces.ViewChangeMessage{message1, message2, message3}
	require.ElementsMatch(t, actualViewChangeMessages, expectedMessages, "stored view-change messages should match the fetched view-change messages")
}

func TestLatestPreprepare(t *testing.T) {
	var s interfaces.Storage = storage.NewInMemoryStorage()
	instanceId := primitives.InstanceId(rand.Uint64())
	blockHeight := primitives.BlockHeight(rand.Uint64())
	senderId1 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId2 := primitives.MemberId(strconv.Itoa(rand.Int()))
	keyManager1 := mocks.NewMockKeyManager(senderId1)
	keyManager2 := mocks.NewMockKeyManager(senderId2)
	block := mocks.ABlock(interfaces.GenesisBlock)

	preprepareMessageOnView3 := builders.APreprepareMessage(instanceId, keyManager1, senderId1, blockHeight, 3, block)
	preprepareMessageOnView2 := builders.APreprepareMessage(instanceId, keyManager2, senderId2, blockHeight, 2, block)

	s.StorePreprepare(preprepareMessageOnView3)
	s.StorePreprepare(preprepareMessageOnView2)

	actualLatestPreprepareMessage, _ := s.GetLatestPreprepare(blockHeight)

	require.Equal(t, actualLatestPreprepareMessage, preprepareMessageOnView3, "fetching preprepare should return the latest preprepare")
}

func TestDuplicatePreprepare(t *testing.T) {
	var s interfaces.Storage = storage.NewInMemoryStorage()
	instanceId := primitives.InstanceId(rand.Uint64())
	block := mocks.ABlock(interfaces.GenesisBlock)
	memberId := primitives.MemberId("Member Id")
	keyManager := mocks.NewMockKeyManager(memberId)
	ppm := builders.APreprepareMessage(instanceId, keyManager, memberId, 1, 1, block)

	firstTime := s.StorePreprepare(ppm)
	require.True(t, firstTime, "StorePreprepare() returns true if storing a new value ")

	secondTime := s.StorePreprepare(ppm)
	require.False(t, secondTime, "StorePreprepare() returns false if trying to store a value that already exists")
}

func TestDuplicatePrepare(t *testing.T) {
	var s interfaces.Storage = storage.NewInMemoryStorage()
	instanceId := primitives.InstanceId(rand.Uint64())
	blockHeight := primitives.BlockHeight(rand.Uint64())
	view := primitives.View(rand.Uint64())
	senderId1 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId2 := primitives.MemberId(strconv.Itoa(rand.Int()))
	sender1KeyManager := mocks.NewMockKeyManager(senderId1)
	sender2KeyManager := mocks.NewMockKeyManager(senderId2)
	block := mocks.ABlock(interfaces.GenesisBlock)
	p1 := builders.APrepareMessage(instanceId, sender1KeyManager, senderId1, blockHeight, view, block)
	p2 := builders.APrepareMessage(instanceId, sender2KeyManager, senderId2, blockHeight, view, block)

	firstTime := s.StorePrepare(p1)
	require.True(t, firstTime, "StorePrepare() returns true if storing a new value (1 of 2)")

	secondTime := s.StorePrepare(p2)
	require.True(t, secondTime, "StorePrepare() returns true if storing a new value (2 of 2)")

	thirdTime := s.StorePrepare(p2)
	require.False(t, thirdTime, "StorePrepare() returns false if trying to store a value that already exists")
}

func TestDuplicateCommit(t *testing.T) {
	var s interfaces.Storage = storage.NewInMemoryStorage()
	instanceId := primitives.InstanceId(rand.Uint64())
	blockHeight := primitives.BlockHeight(rand.Uint64())
	view := primitives.View(rand.Uint64())
	senderId1 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId2 := primitives.MemberId(strconv.Itoa(rand.Int()))
	sender1KeyManager := mocks.NewMockKeyManager(senderId1)
	sender2KeyManager := mocks.NewMockKeyManager(senderId2)
	block := mocks.ABlock(interfaces.GenesisBlock)

	c1 := builders.ACommitMessage(instanceId, sender1KeyManager, senderId1, blockHeight, view, block, 0)
	c2 := builders.ACommitMessage(instanceId, sender2KeyManager, senderId2, blockHeight, view, block, 0)

	firstTime := s.StoreCommit(c1)
	require.True(t, firstTime, "StoreCommit() returns true if storing a new value (1 of 2)")

	secondTime := s.StoreCommit(c2)
	require.True(t, secondTime, "StoreCommit() returns true if storing a new value (2 of 2)")

	thirdTime := s.StoreCommit(c2)
	require.False(t, thirdTime, "StoreCommit() returns false if trying to store a value that already exists")

}

func TestDuplicateViewChange(t *testing.T) {
	var s interfaces.Storage = storage.NewInMemoryStorage()
	instanceId := primitives.InstanceId(rand.Uint64())
	blockHeight := primitives.BlockHeight(rand.Uint64())
	view := primitives.View(rand.Uint64())
	senderId1 := primitives.MemberId(strconv.Itoa(rand.Int()))
	senderId2 := primitives.MemberId(strconv.Itoa(rand.Int()))
	sender1KeyManager := mocks.NewMockKeyManager(senderId1)
	sender2KeyManager := mocks.NewMockKeyManager(senderId2)
	vc1 := builders.AViewChangeMessage(instanceId, sender1KeyManager, senderId1, blockHeight, view, nil)
	vc2 := builders.AViewChangeMessage(instanceId, sender2KeyManager, senderId2, blockHeight, view, nil)

	firstTime := s.StoreViewChange(vc1)
	require.True(t, firstTime, "StoreViewChange() returns true if storing a new value (1 of 2)")

	secondTime := s.StoreViewChange(vc2)
	require.True(t, secondTime, "StoreViewChange() returns true if storing a new value (2 of 2)")

	thirdTime := s.StoreViewChange(vc2)
	require.False(t, thirdTime, "StoreViewChange() returns false if trying to store a value that already exists")

}

func TestClearBlockHeightLogs(t *testing.T) {
	var s interfaces.Storage = storage.NewInMemoryStorage()
	instanceId := primitives.InstanceId(rand.Uint64())
	blockHeight := primitives.BlockHeight(rand.Uint64())
	view := primitives.View(rand.Uint64())
	block := mocks.ABlock(interfaces.GenesisBlock)
	blockHash := mocks.CalculateBlockHash(block)
	memberId := primitives.MemberId("Member Id")
	keyManager := mocks.NewMockKeyManager(memberId)

	ppMsg := builders.APreprepareMessage(instanceId, keyManager, memberId, blockHeight, view, block)
	pMsg := builders.APrepareMessage(instanceId, keyManager, memberId, blockHeight, view, block)
	cMsg := builders.ACommitMessage(instanceId, keyManager, memberId, blockHeight, view, block, 0)
	vcMsg := builders.AViewChangeMessage(instanceId, keyManager, memberId, blockHeight, view, nil)

	s.StorePreprepare(ppMsg)
	s.StorePrepare(pMsg)
	s.StoreCommit(cMsg)
	s.StoreViewChange(vcMsg)

	actualPP, _ := s.GetPreprepareMessage(blockHeight, view)
	actualP, _ := s.GetPrepareMessages(blockHeight, view, blockHash)
	actualC, _ := s.GetCommitMessages(blockHeight, view, blockHash)
	actualVC, _ := s.GetViewChangeMessages(blockHeight, view)
	require.Equal(t, actualPP, ppMsg, "stored preprepare message should match the fetched preprepare message")
	require.Equal(t, 1, len(actualP), "Length of GetPrepareMessages() result array should be 1")
	require.Equal(t, 1, len(actualC), "Length of GetCommitMessages() result array should be 1")
	require.Equal(t, 1, len(actualVC), "Length of GetViewChangeMessages() result array should be 1")

	s.ClearBlockHeightLogs(blockHeight)

	actualPP, _ = s.GetPreprepareMessage(blockHeight, view)
	actualP, _ = s.GetPrepareMessages(blockHeight, view, blockHash)
	actualC, _ = s.GetCommitMessages(blockHeight, view, blockHash)
	actualVC, _ = s.GetViewChangeMessages(blockHeight, view)

	require.Nil(t, actualPP, "GetPreprepareMessage() should return nil after ClearBlockHeightLogs()")
	require.Equal(t, 0, len(actualP), "Length of GetPrepareMessages() result array should be 0")
	require.Equal(t, 0, len(actualC), "Length of GetCommitMessages() result array should be 0")
	require.Equal(t, 0, len(actualVC), "Length of GetViewChangeMessages() result array should be 0")
}

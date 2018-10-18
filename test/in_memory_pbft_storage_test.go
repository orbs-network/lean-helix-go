package test

import (
	"bytes"
	lh "github.com/orbs-network/lean-helix-go"
	. "github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

// TODO TestClearAllStorageDataAfterCallingClearTermLogs
// Ideally Messages should be mocked but it's too much code so using real messages
// TODO 18-OCT-18 go over TS code and rewrite this file

func TestStorePreprepareReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {

	myStorage := lh.NewInMemoryStorage()
	height := BlockHeight(math.Floor(rand.Float64() * 1000))
	view := View(math.Floor(rand.Float64() * 1000))
	block := builders.CreateBlock(builders.GenesisBlock)
	keyManager := builders.NewMockKeyManager(Ed25519PublicKey("PK"))
	mf := lh.NewMessageFactory(keyManager)
	ppm := mf.CreatePreprepareMessage(height, view, block)

	firstTime := myStorage.StorePreprepare(ppm)
	require.True(t, firstTime, "StorePreprepare() returns true if storing a new value ")

	secondTime := myStorage.StorePreprepare(ppm)
	require.False(t, secondTime, "StorePreprepare() returns false if trying to store a value that already exists")
}

func TestClearAllStorageDataAfterCallingClearTermLogs(t *testing.T) {

	myStorage := lh.NewInMemoryStorage()
	height := BlockHeight(math.Floor(rand.Float64() * 1000))
	view := View(math.Floor(rand.Float64() * 1000))
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := builders.CalculateBlockHash(block)
	senderId := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	keyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId))
	msgFactory := lh.NewMessageFactory(keyManager)
	myStorage.StorePreprepare(msgFactory.CreatePreprepareMessage(height, view, block))
	myStorage.StorePrepare(msgFactory.CreatePrepareMessage(height, view, blockHash))
	myStorage.StoreCommit(msgFactory.CreateCommitMessage(height, view, blockHash))
	myStorage.StoreViewChange(msgFactory.CreateViewChangeMessage(height, view, nil))

	pp, _ := myStorage.GetPreprepare(height, view)
	ps, _ := myStorage.GetPrepares(height, view, blockHash)
	require.NotNil(t, pp, "GetPreprepare() should return the store preprepare message")
	require.Equal(t, 1, len(ps), "Length of GetPrepares() result array should be 1")
	require.Equal(t, 1, len(myStorage.GetCommitSendersPKs(height, view, blockHash)), "Length of GetCommitSendersPKs() result array should be 1")
	require.Equal(t, 1, len(myStorage.GetViewChangeMessages(height, view, 0)), "Length of GetViewChangeMessages() result array should be 1")

	myStorage.ClearTermLogs(height)

	pp, _ = myStorage.GetPreprepare(height, view)
	ps, _ = myStorage.GetPrepares(height, view, blockHash)

	require.Nil(t, pp, "GetPreprepare() should return nil after ClearTermLogs()")
	require.Equal(t, 0, len(ps), "Length of GetPrepares() result array should be 0")
	require.Equal(t, 0, len(myStorage.GetCommitSendersPKs(height, view, blockHash)), "Length of GetCommitSendersPKs() result array should be 0")
	require.Equal(t, 0, len(myStorage.GetViewChangeMessages(height, view, 0)), "Length of GetViewChangeMessages() result array should be 0")
}

// TODO func TestStorePrePrepareInStorage
// TODO Do we need TestStorePrePrepareInStorage(t *testing.T) ?

func TestStorePrepareInStorage(t *testing.T) {
	myStorage := lh.NewInMemoryStorage()
	height1 := BlockHeight(math.Floor(rand.Float64() * 1000))
	height2 := BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := View(math.Floor(rand.Float64() * 1000))
	view2 := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId2))
	sender3KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId3))
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(builders.GenesisBlock)
	block1Hash := builders.CalculateBlockHash(block1)
	block2Hash := builders.CalculateBlockHash(block2)
	sender1MsgFactory := lh.NewMessageFactory(sender1KeyManager)
	sender2MsgFactory := lh.NewMessageFactory(sender2KeyManager)
	sender3MsgFactory := lh.NewMessageFactory(sender3KeyManager)
	myStorage.StorePrepare(sender1MsgFactory.CreatePrepareMessage(height1, view1, block1Hash))
	myStorage.StorePrepare(sender2MsgFactory.CreatePrepareMessage(height1, view1, block1Hash))
	myStorage.StorePrepare(sender2MsgFactory.CreatePrepareMessage(height1, view1, block2Hash))
	myStorage.StorePrepare(sender3MsgFactory.CreatePrepareMessage(height1, view2, block1Hash))
	myStorage.StorePrepare(sender3MsgFactory.CreatePrepareMessage(height2, view1, block2Hash))

	expected := []Ed25519PublicKey{senderId1, senderId2}
	actual := myStorage.GetPrepareSendersPKs(height1, view1, block1Hash)
	require.ElementsMatch(t, expected, actual, "Storage stores unique PrePrepare values")
}

func TestStoreCommitInStorage(t *testing.T) {
	myStorage := lh.NewInMemoryStorage()
	height1 := BlockHeight(math.Floor(rand.Float64() * 1000))
	height2 := BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := View(math.Floor(rand.Float64() * 1000))
	view2 := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId2))
	sender3KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId3))
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(builders.GenesisBlock)
	block1Hash := builders.CalculateBlockHash(block1)
	block2Hash := builders.CalculateBlockHash(block2)
	sender1MsgFactory := lh.NewMessageFactory(sender1KeyManager)
	sender2MsgFactory := lh.NewMessageFactory(sender2KeyManager)
	sender3MsgFactory := lh.NewMessageFactory(sender3KeyManager)
	myStorage.StoreCommit(sender1MsgFactory.CreateCommitMessage(height1, view1, block1Hash))
	myStorage.StoreCommit(sender2MsgFactory.CreateCommitMessage(height1, view1, block1Hash))
	myStorage.StoreCommit(sender2MsgFactory.CreateCommitMessage(height1, view1, block2Hash))
	myStorage.StoreCommit(sender3MsgFactory.CreateCommitMessage(height1, view2, block1Hash))
	myStorage.StoreCommit(sender3MsgFactory.CreateCommitMessage(height2, view1, block2Hash))

	expected := []Ed25519PublicKey{senderId1, senderId2}
	actual := myStorage.GetCommitSendersPKs(height1, view1, block1Hash)
	require.ElementsMatch(t, expected, actual, "Storage stores unique PrePrepare values")
}

func TestStorePrepareReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {
	myStorage := lh.NewInMemoryStorage()
	height := BlockHeight(math.Floor(rand.Float64() * 1000))
	view := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId2))
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := builders.CalculateBlockHash(block)
	sender1MsgFactory := lh.NewMessageFactory(sender1KeyManager)
	sender2MsgFactory := lh.NewMessageFactory(sender2KeyManager)
	p1 := sender1MsgFactory.CreatePrepareMessage(height, view, blockHash)
	p2 := sender2MsgFactory.CreatePrepareMessage(height, view, blockHash)

	firstTime := myStorage.StorePrepare(p1)
	require.True(t, firstTime, "StorePrepare() returns true if storing a new value (1 of 2)")

	secondTime := myStorage.StorePrepare(p2)
	require.True(t, secondTime, "StorePrepare() returns true if storing a new value (2 of 2)")

	thirdTime := myStorage.StorePrepare(p2)
	require.False(t, thirdTime, "StorePrepare() returns false if trying to store a value that already exists")
}

func TestStoreCommitReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {
	myStorage := lh.NewInMemoryStorage()
	height := BlockHeight(math.Floor(rand.Float64() * 1000))
	view := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId2))
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := builders.CalculateBlockHash(block)
	sender1MsgFactory := lh.NewMessageFactory(sender1KeyManager)
	sender2MsgFactory := lh.NewMessageFactory(sender2KeyManager)

	c1 := sender1MsgFactory.CreateCommitMessage(height, view, blockHash)
	c2 := sender2MsgFactory.CreateCommitMessage(height, view, blockHash)

	firstTime := myStorage.StoreCommit(c1)
	require.True(t, firstTime, "StoreCommit() returns true if storing a new value (1 of 2)")

	secondTime := myStorage.StoreCommit(c2)
	require.True(t, secondTime, "StoreCommit() returns true if storing a new value (2 of 2)")

	thirdTime := myStorage.StoreCommit(c2)
	require.False(t, thirdTime, "StoreCommit() returns false if trying to store a value that already exists")

}

func TestStoreViewChangeReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {
	myStorage := lh.NewInMemoryStorage()
	height := BlockHeight(math.Floor(rand.Float64() * 1000))
	view := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId2))
	sender1MsgFactory := lh.NewMessageFactory(sender1KeyManager)
	sender2MsgFactory := lh.NewMessageFactory(sender2KeyManager)
	vc1 := sender1MsgFactory.CreateViewChangeMessage(height, view, nil)
	vc2 := sender2MsgFactory.CreateViewChangeMessage(height, view, nil)

	firstTime := myStorage.StoreViewChange(vc1)
	require.True(t, firstTime, "StoreViewChange() returns true if storing a new value (1 of 2)")

	secondTime := myStorage.StoreViewChange(vc2)
	require.True(t, secondTime, "StoreViewChange() returns true if storing a new value (2 of 2)")

	thirdTime := myStorage.StoreViewChange(vc2)
	require.False(t, thirdTime, "StoreViewChange() returns false if trying to store a value that already exists")

}

// Proofs

func TestStoreAndGetViewChangeProof(t *testing.T) {
	myStorage := lh.NewInMemoryStorage()
	height1 := BlockHeight(math.Floor(rand.Float64() * 1000))
	height2 := BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := View(math.Floor(rand.Float64() * 1000))
	senderId1 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := Ed25519PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId2))
	sender3KeyManager := builders.NewMockKeyManager(Ed25519PublicKey(senderId3))
	sender1MsgFactory := lh.NewMessageFactory(sender1KeyManager)
	sender2MsgFactory := lh.NewMessageFactory(sender2KeyManager)
	sender3MsgFactory := lh.NewMessageFactory(sender3KeyManager)
	vcs := make([]lh.ViewChangeMessage, 0, 4)
	vcs = append(vcs, sender1MsgFactory.CreateViewChangeMessage(height1, view1, nil))
	vcs = append(vcs, sender2MsgFactory.CreateViewChangeMessage(height1, view1, nil))
	vcs = append(vcs, sender3MsgFactory.CreateViewChangeMessage(height1, view1, nil))
	vcs = append(vcs, sender3MsgFactory.CreateViewChangeMessage(height2, view1, nil))
	for _, k := range vcs {
		myStorage.StoreViewChange(k)
	}
	f := 1
	actual := myStorage.GetViewChangeMessages(height1, view1, f)
	expected := 2*f + 1                                                     // TODO why this?
	require.Equal(t, expected, len(actual), "return the view-change proof") // TODO bad explanation!
}

func compPrepareProof(t *testing.T, a, b lh.PreparedProof, msg string) {
	require.True(t, bytes.Compare(a.Raw(), b.Raw()) == 0)
}

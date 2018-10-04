package test

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

// TODO TestClearAllStorageDataAfterCallingClearTermLogs

func TestClearAllStorageDataAfterCallingClearTermLogs(t *testing.T) {

	myStorage := lh.NewInMemoryStorage()
	height := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := lh.View(math.Floor(rand.Float64() * 1000))
	block := builders.CreateBlock(builders.GenesisBlock)
	blockHash := builders.CalculateBlockHash(block)
	senderId := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	keyManager := builders.NewMockKeyManager(lh.PublicKey(senderId))
	msgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, keyManager)
	myStorage.StorePreprepare(msgFactory.CreatePreprepareMessage(height, view, block))
	myStorage.StorePrepare(msgFactory.CreatePrepareMessage(height, view, block))
	myStorage.StoreCommit(msgFactory.CreateCommitMessage(height, view, block))
	myStorage.StoreViewChange(msgFactory.CreateViewChangeMessage(height, view, nil, nil))

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
	height1 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	height2 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := lh.View(math.Floor(rand.Float64() * 1000))
	view2 := lh.View(math.Floor(rand.Float64() * 1000))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId2))
	sender3KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId3))
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(builders.GenesisBlock)
	block1Hash := builders.CalculateBlockHash(block1)
	sender1MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender2KeyManager)
	sender3MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender3KeyManager)
	myStorage.StorePrepare(sender1MsgFactory.CreatePrepareMessage(height1, view1, block1))
	myStorage.StorePrepare(sender2MsgFactory.CreatePrepareMessage(height1, view1, block1))
	myStorage.StorePrepare(sender2MsgFactory.CreatePrepareMessage(height1, view1, block2))
	myStorage.StorePrepare(sender3MsgFactory.CreatePrepareMessage(height1, view2, block1))
	myStorage.StorePrepare(sender3MsgFactory.CreatePrepareMessage(height2, view1, block2))

	expected := []lh.PublicKey{senderId1, senderId2}
	actual := myStorage.GetPrepareSendersPKs(height1, view1, block1Hash)
	require.ElementsMatch(t, expected, actual, "Storage stores unique PrePrepare values")
}

func TestStoreCommitInStorage(t *testing.T) {
	myStorage := lh.NewInMemoryStorage()
	height1 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	height2 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := lh.View(math.Floor(rand.Float64() * 1000))
	view2 := lh.View(math.Floor(rand.Float64() * 1000))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId2))
	sender3KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId3))
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(builders.GenesisBlock)
	block1Hash := builders.CalculateBlockHash(block1)
	sender1MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender2KeyManager)
	sender3MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender3KeyManager)
	myStorage.StoreCommit(sender1MsgFactory.CreateCommitMessage(height1, view1, block1))
	myStorage.StoreCommit(sender2MsgFactory.CreateCommitMessage(height1, view1, block1))
	myStorage.StoreCommit(sender2MsgFactory.CreateCommitMessage(height1, view1, block2))
	myStorage.StoreCommit(sender3MsgFactory.CreateCommitMessage(height1, view2, block1))
	myStorage.StoreCommit(sender3MsgFactory.CreateCommitMessage(height2, view1, block2))

	expected := []lh.PublicKey{senderId1, senderId2}
	actual := myStorage.GetCommitSendersPKs(height1, view1, block1Hash)
	require.ElementsMatch(t, expected, actual, "Storage stores unique PrePrepare values")
}

func TestStorePreprepareReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {

	myStorage := lh.NewInMemoryStorage()
	height := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := lh.View(math.Floor(rand.Float64() * 1000))
	block := builders.CreateBlock(builders.GenesisBlock)
	keyManager := builders.NewMockKeyManager(lh.PublicKey("PK"))
	mf := builders.NewMockMessageFactory(builders.CalculateBlockHash, keyManager)
	ppm := mf.CreatePreprepareMessage(height, view, block)

	firstTime := myStorage.StorePreprepare(ppm)
	require.True(t, firstTime, "StorePreprepare() returns true if storing a new value ")

	secondTime := myStorage.StorePreprepare(ppm)
	require.False(t, secondTime, "StorePreprepare() returns false if trying to store a value that already exists")
}

func TestStorePrepareReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {
	myStorage := lh.NewInMemoryStorage()
	height := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := lh.View(math.Floor(rand.Float64() * 1000))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId2))
	block := builders.CreateBlock(builders.GenesisBlock)
	sender1MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender2KeyManager)
	p1 := sender1MsgFactory.CreatePrepareMessage(height, view, block)
	p2 := sender2MsgFactory.CreatePrepareMessage(height, view, block)

	firstTime := myStorage.StorePrepare(p1)
	require.True(t, firstTime, "StorePrepare() returns true if storing a new value (1 of 2)")

	secondTime := myStorage.StorePrepare(p2)
	require.True(t, secondTime, "StorePrepare() returns true if storing a new value (2 of 2)")

	thirdTime := myStorage.StorePrepare(p2)
	require.False(t, thirdTime, "StorePrepare() returns false if trying to store a value that already exists")
}

func TestStoreCommitReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {
	myStorage := lh.NewInMemoryStorage()
	height := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := lh.View(math.Floor(rand.Float64() * 1000))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId2))
	block := builders.CreateBlock(builders.GenesisBlock)
	sender1MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender2KeyManager)

	c1 := sender1MsgFactory.CreateCommitMessage(height, view, block)
	c2 := sender2MsgFactory.CreateCommitMessage(height, view, block)

	firstTime := myStorage.StoreCommit(c1)
	require.True(t, firstTime, "StoreCommit() returns true if storing a new value (1 of 2)")

	secondTime := myStorage.StoreCommit(c2)
	require.True(t, secondTime, "StoreCommit() returns true if storing a new value (2 of 2)")

	thirdTime := myStorage.StoreCommit(c2)
	require.False(t, thirdTime, "StoreCommit() returns false if trying to store a value that already exists")

}

func TestStoreViewChangeReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {
	myStorage := lh.NewInMemoryStorage()
	height := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := lh.View(math.Floor(rand.Float64() * 1000))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId2))
	sender1MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender2KeyManager)
	vc1 := sender1MsgFactory.CreateViewChangeMessage(height, view, nil, nil)
	vc2 := sender2MsgFactory.CreateViewChangeMessage(height, view, nil, nil)

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
	height1 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	height2 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := lh.View(math.Floor(rand.Float64() * 1000))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId2))
	sender3KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId3))
	sender1MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender2KeyManager)
	sender3MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender3KeyManager)
	vcs := make([]lh.ViewChangeMessage, 0, 4)
	vcs = append(vcs, sender1MsgFactory.CreateViewChangeMessage(height1, view1, nil, nil))
	vcs = append(vcs, sender2MsgFactory.CreateViewChangeMessage(height1, view1, nil, nil))
	vcs = append(vcs, sender3MsgFactory.CreateViewChangeMessage(height1, view1, nil, nil))
	vcs = append(vcs, sender3MsgFactory.CreateViewChangeMessage(height2, view1, nil, nil))
	for _, k := range vcs {
		myStorage.StoreViewChange(k)
	}
	f := 1
	actual := myStorage.GetViewChangeMessages(height1, view1, f)
	expected := 2*f + 1                                                     // TODO why this?
	require.Equal(t, expected, len(actual), "return the view-change proof") // TODO bad explanation!
}

func compPrepareProof(t *testing.T, a, b lh.PreparedProof, msg string) {
	require.Equal(t, a.PreprepareMessage(), b.PreprepareMessage(), msg)
	require.ElementsMatch(t, a.PrepareMessages(), b.PrepareMessages(), msg)
}

// from describe("Prepared")
func TestPrepared(t *testing.T) {
	// init here
	fmt.Println("TestPrepared")
	height := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := lh.View(math.Floor(rand.Float64() * 1000))
	leaderId := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	leaderKeyManager := builders.NewMockKeyManager(lh.PublicKey(leaderId))
	sender1KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId1))
	sender2KeyManager := builders.NewMockKeyManager(lh.PublicKey(senderId2))
	block := builders.CreateBlock(builders.GenesisBlock)
	leaderMsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, leaderKeyManager)
	sender1MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := builders.NewMockMessageFactory(builders.CalculateBlockHash, sender2KeyManager)
	ppm := leaderMsgFactory.CreatePreprepareMessage(height, view, block)
	pm1 := sender1MsgFactory.CreatePrepareMessage(height, view, block)
	pm2 := sender2MsgFactory.CreatePrepareMessage(height, view, block)
	f := 1

	// TODO This "TestStoreAndGetPrepareProof" test will always PASS if "TestReturnPreparedProofWithHighestView" below passes, consider deleting
	t.Run("TestStoreAndGetPrepareProof", func(t *testing.T) {
		myStorage := lh.NewInMemoryStorage()
		myStorage.StorePreprepare(ppm)
		myStorage.StorePrepare(pm2)
		myStorage.StorePrepare(pm1)
		expectedProof := lh.CreatePreparedProof(ppm, []lh.PrepareMessage{pm1, pm2})

		actualProof, _ := myStorage.GetLatestPrepared(height, f)
		compPrepareProof(t, actualProof, expectedProof, "return a prepared proof generated by the PPM and PMs in storage")
	})

	t.Run("TestReturnPreparedProofWithHighestView", func(t *testing.T) {
		myStorage := lh.NewInMemoryStorage()
		ppm10 := leaderMsgFactory.CreatePreprepareMessage(1, 10, block)
		pm10a := sender1MsgFactory.CreatePrepareMessage(1, 10, block)
		pm10b := sender2MsgFactory.CreatePrepareMessage(1, 10, block)

		ppm20 := leaderMsgFactory.CreatePreprepareMessage(1, 20, block)
		pm20a := sender1MsgFactory.CreatePrepareMessage(1, 20, block)
		pm20b := sender2MsgFactory.CreatePrepareMessage(1, 20, block)

		ppm30 := leaderMsgFactory.CreatePreprepareMessage(1, 30, block)
		pm30a := sender1MsgFactory.CreatePrepareMessage(1, 30, block)
		pm30b := sender2MsgFactory.CreatePrepareMessage(1, 30, block)

		myStorage.StorePreprepare(ppm10)
		myStorage.StorePrepare(pm10a)
		myStorage.StorePrepare(pm10b)

		myStorage.StorePreprepare(ppm20)
		myStorage.StorePrepare(pm20a)
		myStorage.StorePrepare(pm20b)

		myStorage.StorePreprepare(ppm30)
		myStorage.StorePrepare(pm30a)
		myStorage.StorePrepare(pm30b)

		actual, _ := myStorage.GetLatestPrepared(1, 1)
		require.Equal(t, actual.PreprepareMessage().SignedHeader().View(), lh.View(30), "View of preprepared message should be 30 (highest for this height)")
		require.Equal(t, actual.PrepareMessages()[0].SignedHeader().View(), lh.View(30), "View of prepared message #1 should be 30 (highest for this height)")
		require.Equal(t, actual.PrepareMessages()[1].SignedHeader().View(), lh.View(30), "View of prepared message #2 should be 30 (highest for this height)")
	})

	t.Run("TestReturnNothingIfNoPrePrepare", func(t *testing.T) {
		myStorage := lh.NewInMemoryStorage()
		myStorage.StorePrepare(pm1)
		myStorage.StorePrepare(pm2)
		_, ok := myStorage.GetLatestPrepared(height, f)
		require.False(t, ok, "Don't return PreparedMessages from latest view if no PrePrepare in storage")
	})

	t.Run("TestReturnNothingIfNoPrepares", func(t *testing.T) {
		myStorage := lh.NewInMemoryStorage()
		myStorage.StorePreprepare(ppm)
		_, ok := myStorage.GetLatestPrepared(height, f)
		require.False(t, ok, "Don't return PreparedMessages from latest view if no Prepare in storage")
	})

	t.Run("TestReturnNothingIfNotEnoughPrepares", func(t *testing.T) {
		myStorage := lh.NewInMemoryStorage()
		myStorage.StorePreprepare(ppm)
		myStorage.StorePrepare(pm1)
		_, ok := myStorage.GetLatestPrepared(height, f)
		require.False(t, ok, "Don't return PreparedMessages from latest view if not enough Prepares in storage (# Prepares < 2*f)")
	})
}

// TODO GetLatestPrepared() should initially be here as in TS code but later moved out, because it contains algo logic (it checks something with 2*f))

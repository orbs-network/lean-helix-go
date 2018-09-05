package test

import (
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/orbs-network/lean-helix-go/test/inmemoryblockchain"
	"github.com/orbs-network/lean-helix-go/types"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

// TODO TestClearAllStorageDataAfterCallingClearTermLogs

/*
func TestClearAllStorageDataAfterCallingClearTermLogs(t *testing.T) {

	//const storage = new InMemoryPBFTStorage(logger)

	myStorage := storage.NewInMemoryPBFTStorage()
	term := math.Floor(rand.Int() * 1000)
	view := math.Floor(rand.Int() * 1000)
	block := builders.CreateBlock(builders.CreateGenesisBlock())

	// TODO: This requires orbs-network-go/crypto which cannot be a dependency
	blockHash := digest.CalcTransactionsBlockHash(block)
	keyManager := keymanager.NewMockKeyManager([]byte("PK"), [][]byte{})

	prepreparePayload := CreatePrePrepareMessage(keyManager, term, view, block)
	preparePayload := CreatePrepareMessage(keyManager, term, view, block)
	commitPayload := CreateCommitMessage(keyManager, term, view, block)
	viewChangePayload := CreatePayload(keyManager, nil)

	myStorage.StorePrePrepare(term, view, prepreparePayload)
	myStorage.StorePrepare(term, view, preparePayload)
	myStorage.StoreCommit(term, view, commitPayload)
	myStorage.StoreViewChange(term, view, viewChangePayload)

	require.NotNil(t, storage.GetPrePreparePayload(term, view), "GetPrePreparePayload() result is not nil")
	require.Equal(t, 1, len(storage.GetPreparePayloads(term, view, blockHash)), "Length of GetPreparePayloads() result array is 1")
	require.Equal(t, 1, len(storage.GetCommitSendersPublicKeys(term, view, blockHash)), "Length of GetCommitSendersPublicKeys() result array is 1")
	require.Equal(t, 1, len(storage.GetViewChangeProof(term, view, blockHash)), "Length of GetViewChangeProof() result array is 1")

	storage.ClearTermLogs(term)

	require.Nil(t, storage.GetPrePreparePayload(term, view), "GetPrePreparePayload() result is nil")
	require.Equal(t, 0, len(storage.GetPreparePayloads(term, view, blockHash)), "Length of GetPreparePayloads() result array is 0")
	require.Equal(t, 0, len(storage.GetCommitSendersPublicKeys(term, view, blockHash)), "Length of GetCommitSendersPublicKeys() result array is 0")
	require.Nil(t, 1, len(storage.GetViewChangeProof(term, view, blockHash)), "GetViewChangeProof() result is nil")




	//const term = Math.floor(Math.random() * 1000);
	//const view = Math.floor(Math.random() * 1000);
	//const block = aBlock(theGenesisBlock);
	//const blockHash = calculateBlockHash(block);
	//const keyManager: KeyManager = new mockKeyManager("PK");
	//const PPPayload = aPrePreparePayload(keyManager, term, view, block);
	//const PPayload = aPreparePayload(keyManager, term, view, block);
	//const CPayload = aCommitPayload(keyManager, term, view, block);
	//const VCPayload = aPayload(keyManager, {});
	//
	//// storing
	//storage.storePrePrepare(term, view, PPPayload);
	//storage.storePrepare(term, view, PPayload);
	//storage.storeCommit(term, view, CPayload);
	//storage.storeViewChange(term, view, VCPayload);

	//expect(storage.getPrePreparePayload(term, view)).to.not.be.undefined;
	//expect(storage.getPreparePayloads(term, view, blockHash).length).to.equal(1);
	//expect(storage.getCommitSendersPks(term, view, blockHash).length).to.equal(1);
	//expect(storage.getViewChangeProof(term, view, 0).length).to.equal(1);
	//
	//// clearing
	//storage.clearTermLogs(term);
	//
	//expect(storage.getPrePreparePayload(term, view)).to.be.undefined;
	//expect(storage.getPreparePayloads(term, view, blockHash).length).to.equal(0);
	//expect(storage.getCommitSendersPks(term, view, blockHash).length).to.equal(0);
	//expect(storage.getViewChangeProof(term, view, 0)).to.be.undefined;


}

*/

// TODO func TestStorePrePrepareInStorage
// TODO Do we need TestStorePrePrepareInStorage(t *testing.T) ?

func TestStorePrepareInStorage(t *testing.T) {
	myStorage := leanhelix.NewInMemoryPBFTStorage()
	term1 := types.BlockHeight(math.Floor(rand.Float64() * 1000))
	term2 := types.BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := types.ViewCounter(math.Floor(rand.Float64() * 1000))
	view2 := types.ViewCounter(math.Floor(rand.Float64() * 1000))
	senderId1 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId1))
	sender2KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId2))
	sender3KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId3))
	block1 := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	block2 := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	block1Hash := leanhelix.CalculateBlockHash(block1)
	sender1MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender2KeyManager)
	sender3MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender3KeyManager)
	myStorage.StorePrepare(sender1MsgFactory.CreatePrepareMessage(term1, view1, block1))
	myStorage.StorePrepare(sender2MsgFactory.CreatePrepareMessage(term1, view1, block1))
	myStorage.StorePrepare(sender2MsgFactory.CreatePrepareMessage(term1, view1, block2))
	myStorage.StorePrepare(sender3MsgFactory.CreatePrepareMessage(term1, view2, block1))
	myStorage.StorePrepare(sender3MsgFactory.CreatePrepareMessage(term2, view1, block2))

	expected := []types.PublicKey{senderId1, senderId2}
	actual := myStorage.GetPrepareSendersPKs(term1, view1, block1Hash)
	require.ElementsMatch(t, expected, actual, "Storage stores unique PrePrepare values")
}

func TestStoreCommitInStorage(t *testing.T) {
	myStorage := leanhelix.NewInMemoryPBFTStorage()
	term1 := types.BlockHeight(math.Floor(rand.Float64() * 1000))
	term2 := types.BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := types.ViewCounter(math.Floor(rand.Float64() * 1000))
	view2 := types.ViewCounter(math.Floor(rand.Float64() * 1000))
	senderId1 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId1))
	sender2KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId2))
	sender3KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId3))
	block1 := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	block2 := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	block1Hash := leanhelix.CalculateBlockHash(block1)
	sender1MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender2KeyManager)
	sender3MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender3KeyManager)
	myStorage.StoreCommit(sender1MsgFactory.CreateCommitMessage(term1, view1, block1))
	myStorage.StoreCommit(sender2MsgFactory.CreateCommitMessage(term1, view1, block1))
	myStorage.StoreCommit(sender2MsgFactory.CreateCommitMessage(term1, view1, block2))
	myStorage.StoreCommit(sender3MsgFactory.CreateCommitMessage(term1, view2, block1))
	myStorage.StoreCommit(sender3MsgFactory.CreateCommitMessage(term2, view1, block2))

	expected := []types.PublicKey{senderId1, senderId2}
	actual := myStorage.GetCommitSendersPKs(term1, view1, block1Hash)
	require.ElementsMatch(t, expected, actual, "Storage stores unique PrePrepare values")
}

func TestStorePreprepareReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {

	myStorage := leanhelix.NewInMemoryPBFTStorage()
	term := types.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := types.ViewCounter(math.Floor(rand.Float64() * 1000))
	block := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	keyManager := leanhelix.NewMockKeyManager(types.PublicKey("PK"))
	mf := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, keyManager)
	ppContent := mf.CreatePreprepareMessage(term, view, block)

	firstTime := myStorage.StorePrePrepare(ppContent)
	require.True(t, firstTime, "StorePrePrepare() returns true if storing a new value ")

	secondTime := myStorage.StorePrePrepare(ppContent)
	require.False(t, secondTime, "StorePrePrepare() returns false if trying to store a value that already exists")
}

func TestStorePrepareReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {
	myStorage := leanhelix.NewInMemoryPBFTStorage()
	term := types.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := types.ViewCounter(math.Floor(rand.Float64() * 1000))
	senderId1 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId1))
	sender2KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId2))
	block := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	sender1MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender2KeyManager)
	prepareMessage1 := sender1MsgFactory.CreatePrepareMessage(term, view, block)
	prepareMessage2 := sender2MsgFactory.CreatePrepareMessage(term, view, block)

	firstTime := myStorage.StorePrepare(prepareMessage1)
	require.True(t, firstTime, "StorePrepare() returns true if storing a new value (1 of 2)")

	secondTime := myStorage.StorePrepare(prepareMessage2)
	require.True(t, secondTime, "StorePrepare() returns true if storing a new value (2 of 2)")

	thirdTime := myStorage.StorePrepare(prepareMessage2)
	require.False(t, thirdTime, "StorePrepare() returns false if trying to store a value that already exists")
}

func TestStoreCommitReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {
	myStorage := leanhelix.NewInMemoryPBFTStorage()
	term := types.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := types.ViewCounter(math.Floor(rand.Float64() * 1000))
	senderId1 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId1))
	sender2KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId2))
	block := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	sender1MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender2KeyManager)

	commitPayload1 := sender1MsgFactory.CreateCommitMessage(term, view, block)
	commitPayload2 := sender2MsgFactory.CreateCommitMessage(term, view, block)

	firstTime := myStorage.StoreCommit(commitPayload1)
	require.True(t, firstTime, "StoreCommit() returns true if storing a new value (1 of 2)")

	secondTime := myStorage.StoreCommit(commitPayload2)
	require.True(t, secondTime, "StoreCommit() returns true if storing a new value (2 of 2)")

	thirdTime := myStorage.StoreCommit(commitPayload2)
	require.False(t, thirdTime, "StoreCommit() returns false if trying to store a value that already exists")

}

// TODO TestStoreViewChangeReturnsTrueIfNewOrFalseIfAlreadyExists

// Proofs

func TestStoreAndGetViewChangeProof(t *testing.T) {
	myStorage := leanhelix.NewInMemoryPBFTStorage()
	term1 := types.BlockHeight(math.Floor(rand.Float64() * 1000))
	term2 := types.BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := types.ViewCounter(math.Floor(rand.Float64() * 1000))
	senderId1 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId1))
	sender2KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId2))
	sender3KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId3))
	sender1MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender2KeyManager)
	sender3MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender3KeyManager)
	vcms := make([]*leanhelix.ViewChangeMessage, 0, 4)
	vcms = append(vcms, sender1MsgFactory.CreateViewChangeMessage(term1, view1, nil))
	vcms = append(vcms, sender2MsgFactory.CreateViewChangeMessage(term1, view1, nil))
	vcms = append(vcms, sender3MsgFactory.CreateViewChangeMessage(term1, view1, nil))
	vcms = append(vcms, sender3MsgFactory.CreateViewChangeMessage(term2, view1, nil))
	for _, k := range vcms {
		myStorage.StoreViewChange(k)
	}
	f := 1
	actual := myStorage.GetViewChangeMessages(term1, view1, f)
	expected := 2*f + 1                                                     // TODO why this?
	require.Equal(t, expected, len(actual), "return the view-change proof") // TODO bad explanation!
}

func compPrepareMessages(t *testing.T, a, b *leanhelix.PreparedMessages, msg string) {
	require.Equal(t, a.PreprepareMessage, b.PreprepareMessage, msg)
	require.ElementsMatch(t, a.PrepareMessages, b.PrepareMessages, msg)
}

// from describe("Prepared")
func TestPrepared(t *testing.T) {
	// init here
	fmt.Println("TestPrepared")
	term := types.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := types.ViewCounter(math.Floor(rand.Float64() * 1000))
	leaderId := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId1 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := types.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	leaderKeyManager := leanhelix.NewMockKeyManager(types.PublicKey(leaderId))
	sender1KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId1))
	sender2KeyManager := leanhelix.NewMockKeyManager(types.PublicKey(senderId2))
	block := builders.CreateBlock(inmemoryblockchain.GenesisBlock)
	leaderMsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, leaderKeyManager)
	sender1MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender1KeyManager)
	sender2MsgFactory := leanhelix.NewMessageFactory(leanhelix.CalculateBlockHash, sender2KeyManager)
	ppm := leaderMsgFactory.CreatePreprepareMessage(term, view, block)
	pm1 := sender1MsgFactory.CreatePrepareMessage(term, view, block)
	pm2 := sender2MsgFactory.CreatePrepareMessage(term, view, block)
	f := 1

	t.Run("TestStoreAndGetPrepareProof", func(t *testing.T) {
		myStorage := leanhelix.NewInMemoryPBFTStorage()
		myStorage.StorePrePrepare(ppm)
		myStorage.StorePrepare(pm2)
		myStorage.StorePrepare(pm1)
		expectedProof := &leanhelix.PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages:   []*leanhelix.PrepareMessage{pm1, pm2},
		}

		actualProof, _ := myStorage.GetLatestPrepared(term, f)
		compPrepareMessages(t, actualProof, expectedProof, "TestStoreAndGetPrepareProof(): return the prepared proof") // TODO bad explanation!
	})

	//t.Run("TestReturnPreparedProofWithHighestView", func(t *testing.T) {
	//	myStorage := storage.NewInMemoryPBFTStorage()
	//	prePrepareMessage10 := builders.CreatePrePrepareMessage(leaderKeyManager, 1, 10, block)
	//	prepareMessage10_1 := builders.CreatePrepareMessage(sender1KeyManager, 1, 10, block)
	//	prepareMessage10_2 := builders.CreatePrepareMessage(sender2KeyManager, 1, 10, block)
	//
	//	prePrepareMessage20 := builders.CreatePrePrepareMessage(leaderKeyManager, 1, 20, block)
	//	prepareMessage20_1 := builders.CreatePrepareMessage(sender1KeyManager, 1, 20, block)
	//	prepareMessage20_2 := builders.CreatePrepareMessage(sender2KeyManager, 1, 20, block)
	//
	//	prePrepareMessage30 := builders.CreatePrePrepareMessage(leaderKeyManager, 1, 30, block)
	//	prepareMessage30_1 := builders.CreatePrepareMessage(sender1KeyManager, 1, 30, block)
	//	prepareMessage30_2 := builders.CreatePrepareMessage(sender2KeyManager, 1, 30, block)
	//
	//	myStorage.StorePrePrepare(prePrepareMessage10)
	//	myStorage.StorePrepare(prepareMessage10_1)
	//	myStorage.StorePrepare(prepareMessage10_2)
	//
	//	myStorage.StorePrePrepare(prePrepareMessage20)
	//	myStorage.StorePrepare(prepareMessage20_1)
	//	myStorage.StorePrepare(prepareMessage20_2)
	//
	//	myStorage.StorePrePrepare(prePrepareMessage30)
	//	myStorage.StorePrepare(prepareMessage30_1)
	//	myStorage.StorePrepare(prepareMessage30_2)
	//
	//	expected := &PreparedMessages{
	//		PreprepareMessage: prePrepareMessage30,
	//		PrepareMessages:   []*PrepareMessage{prepareMessage30_1, prepareMessage30_2},
	//	}
	//	actual, _ := myStorage.GetLatestPrepared(1, 1)
	//	require.ElementsMatch(t, expected, actual, "TestReturnPreparedProofWithHighestView")
	//})

	t.Run("TestReturnNothingIfNoPrePrepare", func(t *testing.T) {
		myStorage := leanhelix.NewInMemoryPBFTStorage()
		myStorage.StorePrepare(pm1)
		myStorage.StorePrepare(pm2)
		_, ok := myStorage.GetLatestPrepared(term, f)
		require.False(t, ok, "Don't return PreparedMessages from latest view if no PrePrepare in storage")
	})

	t.Run("TestReturnNothingIfNoPrepares", func(t *testing.T) {
		myStorage := leanhelix.NewInMemoryPBFTStorage()
		myStorage.StorePrePrepare(ppm)
		_, ok := myStorage.GetLatestPrepared(term, f)
		require.False(t, ok, "Don't return PreparedMessages from latest view if no Prepare in storage")
	})

	t.Run("TestReturnNothingIfNotEnoughPrepares", func(t *testing.T) {
		myStorage := leanhelix.NewInMemoryPBFTStorage()
		myStorage.StorePrePrepare(ppm)
		myStorage.StorePrepare(pm1)
		_, ok := myStorage.GetLatestPrepared(term, f)
		require.False(t, ok, "Don't return PreparedMessages from latest view if not enough Prepares in storage (# Prepares < 2*f)")
	})
}

// TODO func TestStoreAndGetPrepareProof
// TODO func TestReturnHighestPrepareProof
// TODO func TestReturnUndefinedIfNoPreprepare
// TODO func TestReturnUndefinedIfNoPrepares
// TODO func TestReturnUndefinedIfNotEnoughPrepares

// TODO GetLatestPrepared() should initially be here as in TS code but later moved out, because it contains algo logic (it checks something with 2*f))

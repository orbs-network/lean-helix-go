package storage

import (
	"fmt"
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/storage"
	"github.com/orbs-network/lean-helix-go/go/test/builders"
	"github.com/orbs-network/lean-helix-go/go/test/keymanagermock"
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
	myStorage := storage.NewInMemoryPBFTStorage()
	term1 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	term2 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := lh.ViewCounter(math.Floor(rand.Float64() * 1000))
	view2 := lh.ViewCounter(math.Floor(rand.Float64() * 1000))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId1), []lh.PublicKey{})
	sender2KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId2), []lh.PublicKey{})
	sender3KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId3), []lh.PublicKey{})
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(builders.GenesisBlock)
	block1Hash := builders.CalculateBlockHash(block1)
	myStorage.StorePrepare(builders.CreatePrepareMessage(sender1KeyManager, term1, view1, block1))
	myStorage.StorePrepare(builders.CreatePrepareMessage(sender2KeyManager, term1, view1, block1))
	myStorage.StorePrepare(builders.CreatePrepareMessage(sender2KeyManager, term1, view1, block2))
	myStorage.StorePrepare(builders.CreatePrepareMessage(sender3KeyManager, term1, view2, block1))
	myStorage.StorePrepare(builders.CreatePrepareMessage(sender3KeyManager, term2, view1, block2))

	expected := []lh.PublicKey{senderId1, senderId2}
	actual := myStorage.GetPrepareSendersPKs(term1, view1, block1Hash)
	require.ElementsMatch(t, expected, actual, "Storage stores unique PrePrepare values")
}

func TestStoreCommitInStorage(t *testing.T) {
	myStorage := storage.NewInMemoryPBFTStorage()
	term1 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	term2 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := lh.ViewCounter(math.Floor(rand.Float64() * 1000))
	view2 := lh.ViewCounter(math.Floor(rand.Float64() * 1000))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId1), []lh.PublicKey{})
	sender2KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId2), []lh.PublicKey{})
	sender3KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId3), []lh.PublicKey{})
	block1 := builders.CreateBlock(builders.GenesisBlock)
	block2 := builders.CreateBlock(builders.GenesisBlock)
	block1Hash := builders.CalculateBlockHash(block1)
	myStorage.StoreCommit(builders.CreateCommitMessage(sender1KeyManager, term1, view1, block1))
	myStorage.StoreCommit(builders.CreateCommitMessage(sender2KeyManager, term1, view1, block1))
	myStorage.StoreCommit(builders.CreateCommitMessage(sender2KeyManager, term1, view1, block2))
	myStorage.StoreCommit(builders.CreateCommitMessage(sender3KeyManager, term1, view2, block1))
	myStorage.StoreCommit(builders.CreateCommitMessage(sender3KeyManager, term2, view1, block2))

	expected := []lh.PublicKey{senderId1, senderId2}
	actual := myStorage.GetCommitSendersPKs(term1, view1, block1Hash)
	require.ElementsMatch(t, expected, actual, "Storage stores unique PrePrepare values")
}

func TestStorePreprepareReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {

	myStorage := storage.NewInMemoryPBFTStorage()
	term := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := lh.ViewCounter(math.Floor(rand.Float64() * 1000))
	block := builders.CreateBlock(builders.GenesisBlock)
	keyManager := keymanagermock.NewMockKeyManager(lh.PublicKey("PK"), []lh.PublicKey{})
	ppContent := builders.CreatePrePrepareMessage(keyManager, term, view, block)

	firstTime := myStorage.StorePrePrepare(ppContent)
	require.True(t, firstTime, "StorePrePrepare() returns true if storing a new value ")

	secondTime := myStorage.StorePrePrepare(ppContent)
	require.False(t, secondTime, "StorePrePrepare() returns false if trying to store a value that already exists")
}

func TestStorePrepareReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {
	myStorage := storage.NewInMemoryPBFTStorage()
	term := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := lh.ViewCounter(math.Floor(rand.Float64() * 1000))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId1), []lh.PublicKey{})
	sender2KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId2), []lh.PublicKey{})
	block := builders.CreateBlock(builders.GenesisBlock)
	prepareMessage1 := builders.CreatePrepareMessage(sender1KeyManager, term, view, block)
	prepareMessage2 := builders.CreatePrepareMessage(sender2KeyManager, term, view, block)

	firstTime := myStorage.StorePrepare(prepareMessage1)
	require.True(t, firstTime, "StorePrepare() returns true if storing a new value (1 of 2)")

	secondTime := myStorage.StorePrepare(prepareMessage2)
	require.True(t, secondTime, "StorePrepare() returns true if storing a new value (2 of 2)")

	thirdTime := myStorage.StorePrepare(prepareMessage2)
	require.False(t, thirdTime, "StorePrepare() returns false if trying to store a value that already exists")
}

func TestStoreCommitReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {
	myStorage := storage.NewInMemoryPBFTStorage()
	term := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := lh.ViewCounter(math.Floor(rand.Float64() * 1000))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId1), []lh.PublicKey{})
	sender2KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId2), []lh.PublicKey{})
	block := builders.CreateBlock(builders.GenesisBlock)

	commitPayload1 := builders.CreateCommitMessage(sender1KeyManager, term, view, block)
	commitPayload2 := builders.CreateCommitMessage(sender2KeyManager, term, view, block)

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
	myStorage := storage.NewInMemoryPBFTStorage()
	term1 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	term2 := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view1 := lh.ViewCounter(math.Floor(rand.Float64() * 1000))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId3 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	sender1KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId1), []lh.PublicKey{})
	sender2KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId2), []lh.PublicKey{})
	sender3KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId3), []lh.PublicKey{})
	vcms := make([]*lh.ViewChangeMessage, 0, 4)
	vcms = append(vcms, builders.CreateViewChangeMessage(sender1KeyManager, term1, view1, nil))
	vcms = append(vcms, builders.CreateViewChangeMessage(sender2KeyManager, term1, view1, nil))
	vcms = append(vcms, builders.CreateViewChangeMessage(sender3KeyManager, term1, view1, nil))
	vcms = append(vcms, builders.CreateViewChangeMessage(sender3KeyManager, term2, view1, nil))
	for _, k := range vcms {
		myStorage.StoreViewChange(k)
	}
	f := 1
	actual := myStorage.GetViewChangeMessages(term1, view1, f)
	expected := 2*f + 1                                                     // TODO why this?
	require.Equal(t, expected, len(actual), "return the view-change proof") // TODO bad explanation!
}

// from describe("Prepared")
func TestPrepared(t *testing.T) {
	// init here
	fmt.Println("TestPrepared")
	term := lh.BlockHeight(math.Floor(rand.Float64() * 1000))
	view := lh.ViewCounter(math.Floor(rand.Float64() * 1000))
	leaderId := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId1 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	senderId2 := lh.PublicKey(strconv.Itoa(int(math.Floor(rand.Float64() * 1000))))
	leaderKeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(leaderId), []lh.PublicKey{})
	sender1KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId1), []lh.PublicKey{})
	sender2KeyManager := keymanagermock.NewMockKeyManager(lh.PublicKey(senderId2), []lh.PublicKey{})
	block := builders.CreateBlock(builders.GenesisBlock)
	ppm := builders.CreatePrePrepareMessage(leaderKeyManager, term, view, block)
	pm1 := builders.CreatePrepareMessage(sender1KeyManager, term, view, block)
	pm2 := builders.CreatePrepareMessage(sender2KeyManager, term, view, block)
	f := 1

	t.Run("TestStoreAndGetPrepareProof", func(t *testing.T) {
		myStorage := storage.NewInMemoryPBFTStorage()
		myStorage.StorePrePrepare(ppm)
		myStorage.StorePrepare(pm1)
		myStorage.StorePrepare(pm2)
		expectedProof := &lh.PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages:   []*lh.BlockRefMessage{pm1, pm2},
		}

		actualProof, _ := myStorage.GetLatestPrepared(term, f)
		require.Equal(t, expectedProof, actualProof, "TestStoreAndGetPrepareProof(): return the prepared proof") // TODO bad explanation!
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
	//	expected := &lh.PreparedMessages{
	//		PreprepareMessage: prePrepareMessage30,
	//		PrepareMessages:   []*lh.PrepareMessage{prepareMessage30_1, prepareMessage30_2},
	//	}
	//	actual, _ := myStorage.GetLatestPrepared(1, 1)
	//	require.ElementsMatch(t, expected, actual, "TestReturnPreparedProofWithHighestView")
	//})

	t.Run("TestReturnNothingIfNoPrePrepare", func(t *testing.T) {
		myStorage := storage.NewInMemoryPBFTStorage()
		myStorage.StorePrepare(pm1)
		myStorage.StorePrepare(pm2)
		_, ok := myStorage.GetLatestPrepared(term, f)
		require.False(t, ok, "Don't return PreparedMessages from latest view if no PrePrepare in storage")
	})

	t.Run("TestReturnNothingIfNoPrepares", func(t *testing.T) {
		myStorage := storage.NewInMemoryPBFTStorage()
		myStorage.StorePrePrepare(ppm)
		_, ok := myStorage.GetLatestPrepared(term, f)
		require.False(t, ok, "Don't return PreparedMessages from latest view if no Prepare in storage")
	})

	t.Run("TestReturnNothingIfNotEnoughPrepares", func(t *testing.T) {
		myStorage := storage.NewInMemoryPBFTStorage()
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

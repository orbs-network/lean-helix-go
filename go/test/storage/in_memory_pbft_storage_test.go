package storage

import (
	"github.com/orbs-network/lean-helix-go/go/storage"
	"github.com/orbs-network/lean-helix-go/go/test/builders"
	"github.com/orbs-network/lean-helix-go/go/test/keymanagermock"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

/*
func TestClearAllStorageDataAfterCallingClearTermLogs(t *testing.T) {

	//const storage = new InMemoryPBFTStorage(logger)

	myStorage := storage.NewInMemoryPBFTStorage()
	term := math.Floor(rand.Int() * 1000)
	view := math.Floor(rand.Int() * 1000)
	block := builders.CreateBlock(builders.CreateGenesisBlock())

	// TODO: This requires orbs-network-go/crypto which cannot be a dependency
	blockHash := digest.CalcTransactionsBlockHash(block)
	keyManager := keymanager.NewKeyManagerMock([]byte("PK"), [][]byte{})

	prepreparePayload := CreatePrepreparePayload(keyManager, term, view, block)
	preparePayload := CreatePreparePayload(keyManager, term, view, block)
	commitPayload := CreateCommitPayload(keyManager, term, view, block)
	viewChangePayload := CreatePayload(keyManager, nil)

	myStorage.StorePreprepare(term, view, prepreparePayload)
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
	//const keyManager: KeyManager = new KeyManagerMock("PK");
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

func TestStorePreprepareReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {

	myStorage := storage.NewInMemoryPBFTStorage()
	term := uint64(math.Floor(rand.Float64() * 1000))
	view := uint64(math.Floor(rand.Float64() * 1000))
	block := builders.CreateBlock(builders.GenesisBlock)
	keyManager := keymanagermock.NewKeyManagerMock([]byte("PK"), [][]byte{})
	prepreparePayload := builders.CreatePrepreparePayload(keyManager, term, view, block)

	firstTime := myStorage.StorePreprepare(term, view, prepreparePayload)
	require.True(t, firstTime, "StorePreprepare() returns true if storing a new value ")

	secondTime := myStorage.StorePreprepare(term, view, prepreparePayload)
	require.False(t, secondTime, "StorePreprepare() returns false if trying to store a value that already exists")
}

func TestStorePrepareReturnsTrueIfNewOrFalseIfAlreadyExists(t *testing.T) {
	myStorage := storage.NewInMemoryPBFTStorage()
	term := uint64(math.Floor(rand.Float64() * 1000))
	view := uint64(math.Floor(rand.Float64() * 1000))
	senderId1 := string(uint64(math.Floor(rand.Float64() * 1000)))
	senderId2 := string(uint64(math.Floor(rand.Float64() * 1000)))
	sender1KeyManager := keymanagermock.NewKeyManagerMock([]byte(senderId1), [][]byte{})
	sender2KeyManager := keymanagermock.NewKeyManagerMock([]byte(senderId2), [][]byte{})
	block := builders.CreateBlock(builders.GenesisBlock)
	preparePayload1 := builders.CreatePreparePayload(sender1KeyManager, term, view, block)
	preparePayload2 := builders.CreatePreparePayload(sender2KeyManager, term, view, block)

	firstTime := myStorage.StorePrepare(term, view, preparePayload1)
	require.True(t, firstTime, "StorePrepare() returns true if storing a new value (1 of 2)")

	secondTime := myStorage.StorePrepare(term, view, preparePayload2)
	require.True(t, secondTime, "StorePrepare() returns true if storing a new value (2 of 2)")

	thirdTime := myStorage.StorePrepare(term, view, preparePayload2)
	require.False(t, thirdTime, "StorePrepare() returns false if trying to store a value that already exists")

	/*
		const storage = new InMemoryPBFTStorage(logger);
	const term = Math.floor(Math.random() * 1000);
	const view = Math.floor(Math.random() * 1000);
	const senderId1 = Math.floor(Math.random() * 1000).toString();
	const senderId2 = Math.floor(Math.random() * 1000).toString();
	const sender1KeyManager: KeyManager = new KeyManagerMock(senderId1);
	const sender2KeyManager: KeyManager = new KeyManagerMock(senderId2);
	const block = aBlock(theGenesisBlock);
	const firstTime = storage.storePrepare(term, view, aPreparePayload(sender1KeyManager, term, view, block));
	expect(firstTime).to.be.true;
	const secondstime = storage.storePrepare(term, view, aPreparePayload(sender2KeyManager, term, view, block));
	expect(secondstime).to.be.true;
	const thirdTime = storage.storePrepare(term, view, aPreparePayload(sender2KeyManager, term, view, block));
	expect(thirdTime).to.be.false;

	*/

}

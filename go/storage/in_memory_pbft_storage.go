package storage

import (
	"github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/utils"
)

type inMemoryPbftStorage struct {
	// TODO Refactor this mess - in the least create some intermediate types
	preprepareStorage map[uint64]map[uint64]*leanhelix.PrepreparePayload
	prepareStorage    map[uint64]map[uint64]map[string]map[string]*leanhelix.PreparePayload
	commitStorage     map[uint64]map[uint64]map[string]map[string]*leanhelix.CommitPayload
}

func (storage *inMemoryPbftStorage) StorePreprepare(term uint64, view uint64, prepreparePayload *leanhelix.PrepreparePayload) bool {

	views, ok := storage.preprepareStorage[term]
	if !ok {
		views = make(map[uint64]*leanhelix.PrepreparePayload)
		storage.preprepareStorage[term] = views
	}

	_, ok = views[view]
	if ok {
		return false
	}
	views[view] = prepreparePayload
	return true
}

func (storage *inMemoryPbftStorage) StorePrepare(term uint64, view uint64, preparePayload *leanhelix.PreparePayload) bool {
	// pps -> views ->
	views, ok := storage.prepareStorage[term]
	if !ok {
		views = make(map[uint64]map[string]map[string]*leanhelix.PreparePayload)
		storage.prepareStorage[term] = views
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[string]map[string]*leanhelix.PreparePayload)
		views[view] = blockHashes
	}
	key := string(preparePayload.Data.BlockHash)
	senders, ok := blockHashes[key]
	if !ok {
		senders = make(map[string]*leanhelix.PreparePayload)
		blockHashes[key] = senders
	}
	senderPublicKey := string(preparePayload.PublicKey)
	_, ok = senders[senderPublicKey]
	if ok {
		return false
	}
	senders[senderPublicKey] = preparePayload

	utils.Logger.Info("StorePrepare: term=%d view=%d, senderPk=%s, blockHash=%s", term, view, senderPublicKey, key)

	return true
}

func (storage *inMemoryPbftStorage) StoreCommit(term uint64, view uint64, commitPayload *leanhelix.CommitPayload) bool {
	// pps -> views ->
	views, ok := storage.commitStorage[term]
	if !ok {
		views = make(map[uint64]map[string]map[string]*leanhelix.CommitPayload)
		storage.commitStorage[term] = views
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[string]map[string]*leanhelix.CommitPayload)
		views[view] = blockHashes
	}
	key := string(commitPayload.Data.BlockHash)
	senders, ok := blockHashes[key]
	if !ok {
		senders = make(map[string]*leanhelix.CommitPayload)
		blockHashes[key] = senders
	}
	senderPublicKey := string(commitPayload.PublicKey)
	_, ok = senders[senderPublicKey]
	if ok {
		return false
	}
	senders[senderPublicKey] = commitPayload

	utils.Logger.Info("StoreCommit: term=%d view=%d, senderPk=%s, blockHash=%s", term, view, senderPublicKey, key)

	return true

}

func NewInMemoryPBFTStorage() *inMemoryPbftStorage {
	return &inMemoryPbftStorage{
		preprepareStorage: make(map[uint64]map[uint64]*leanhelix.PrepreparePayload),
		prepareStorage:    make(map[uint64]map[uint64]map[string]map[string]*leanhelix.PreparePayload),
		commitStorage:     make(map[uint64]map[uint64]map[string]map[string]*leanhelix.CommitPayload),
	}
}

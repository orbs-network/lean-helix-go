package storage

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/utils"
)

type inMemoryPbftStorage struct {
	// TODO Refactor this mess - in the least create some intermediate types
	preprepareStorage map[lh.BlockHeight]map[lh.ViewCounter]*lh.PrePrepareMessage
	prepareStorage    map[lh.BlockHeight]map[lh.ViewCounter]map[string]map[string]*lh.PrepareMessage
	commitStorage     map[lh.BlockHeight]map[lh.ViewCounter]map[string]map[string]*lh.CommitMessage
}

func (storage *inMemoryPbftStorage) StorePrePrepare(term lh.BlockHeight, view lh.ViewCounter, ppm *lh.PrePrepareMessage) bool {

	views, ok := storage.preprepareStorage[term]
	if !ok {
		views = make(map[lh.ViewCounter]*lh.PrePrepareMessage)
		storage.preprepareStorage[term] = views
	}

	_, ok = views[view]
	if ok {
		return false
	}
	views[view] = ppm
	return true
}

func (storage *inMemoryPbftStorage) StorePrepare(term lh.BlockHeight, view lh.ViewCounter, pp *lh.PrepareMessage) bool {
	// pps -> views ->
	views, ok := storage.prepareStorage[term]
	if !ok {
		views = make(map[lh.ViewCounter]map[string]map[string]*lh.PrepareMessage)
		storage.prepareStorage[term] = views
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[string]map[string]*lh.PrepareMessage)
		views[view] = blockHashes
	}
	key := string(pp.Content.BlockHash)
	senders, ok := blockHashes[key]
	if !ok {
		senders = make(map[string]*lh.PrepareMessage)
		blockHashes[key] = senders
	}
	senderPublicKey := string(pp.SignaturePair.SignerPublicKey)
	_, ok = senders[senderPublicKey]
	if ok {
		return false
	}
	senders[senderPublicKey] = pp

	utils.Logger.Info("StorePrepare: term=%d view=%d, senderPk=%s, blockHash=%s", term, view, senderPublicKey, key)

	return true
}

func (storage *inMemoryPbftStorage) StoreCommit(term lh.BlockHeight, view lh.ViewCounter, cm *lh.CommitMessage) bool {
	// pps -> views ->
	views, ok := storage.commitStorage[term]
	if !ok {
		views = make(map[lh.ViewCounter]map[string]map[string]*lh.CommitMessage)
		storage.commitStorage[term] = views
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[string]map[string]*lh.CommitMessage)
		views[view] = blockHashes
	}
	key := string(cm.Content.BlockHash)
	senders, ok := blockHashes[key]
	if !ok {
		senders = make(map[string]*lh.CommitMessage)
		blockHashes[key] = senders
	}
	senderPublicKey := string(cm.SignaturePair.SignerPublicKey)
	_, ok = senders[senderPublicKey]
	if ok {
		return false
	}
	senders[senderPublicKey] = cm

	utils.Logger.Info("StoreCommit: term=%d view=%d, senderPk=%s, blockHash=%s", term, view, senderPublicKey, key)

	return true

}

func NewInMemoryPBFTStorage() *inMemoryPbftStorage {
	return &inMemoryPbftStorage{
		preprepareStorage: make(map[lh.BlockHeight]map[lh.ViewCounter]*lh.PrePrepareMessage),
		prepareStorage:    make(map[lh.BlockHeight]map[lh.ViewCounter]map[string]map[string]*lh.PrepareMessage),
		commitStorage:     make(map[lh.BlockHeight]map[lh.ViewCounter]map[string]map[string]*lh.CommitMessage),
	}
}

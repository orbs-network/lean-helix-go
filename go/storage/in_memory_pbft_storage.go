package storage

import (
	"github.com/orbs-network/lean-helix-go/go/networkcommunication"
	"github.com/orbs-network/lean-helix-go/go/utils"
)

type inMemoryPbftStorage struct {
	// TODO Refactor this mess - in the least create some intermediate types
	preprepareStorage map[uint64]map[uint64]*networkcommunication.PrepreparePayload
	prepareStorage    map[uint64]map[uint64]map[string]map[string]*networkcommunication.PreparePayload
}

func (storage *inMemoryPbftStorage) StorePreprepare(term uint64, view uint64, prepreparePayload *networkcommunication.PrepreparePayload) bool {

	views, ok := storage.preprepareStorage[term]
	if !ok {
		views = make(map[uint64]*networkcommunication.PrepreparePayload)
		storage.preprepareStorage[term] = views
	}

	_, ok = views[view]
	if ok {
		return false
	}
	views[view] = prepreparePayload
	return true
}

func (storage *inMemoryPbftStorage) StorePrepare(term uint64, view uint64, preparePayload *networkcommunication.PreparePayload) bool {
	// pps -> views ->
	views, ok := storage.prepareStorage[term]
	if !ok {
		views = make(map[uint64]map[string]map[string]*networkcommunication.PreparePayload)
		storage.prepareStorage[term] = views
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[string]map[string]*networkcommunication.PreparePayload)
		views[view] = blockHashes
	}
	key := string(preparePayload.Data.BlockHash)
	senders, ok := blockHashes[key]
	if !ok {
		senders = make(map[string]*networkcommunication.PreparePayload)
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

func NewInMemoryPBFTStorage() *inMemoryPbftStorage {
	return &inMemoryPbftStorage{
		preprepareStorage: make(map[uint64]map[uint64]*networkcommunication.PrepreparePayload),
		prepareStorage:    make(map[uint64]map[uint64]map[string]map[string]*networkcommunication.PreparePayload),
	}
}

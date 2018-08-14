package storage

import (
	"github.com/orbs-network/lean-helix-go/go/networkcommunication"
)

type inMemoryPbftStorage struct {
	preprepareStorage map[uint64]map[uint64]*networkcommunication.PrepreparePayload
}

func (storage *inMemoryPbftStorage) StorePreprepare(term uint64, view uint64, prepreparePayload *networkcommunication.PrepreparePayload) bool {

	termsMap, ok := storage.preprepareStorage[term]
	if !ok {
		termsMap = make(map[uint64]*networkcommunication.PrepreparePayload)
		storage.preprepareStorage[term] = termsMap
	}

	_, ok = termsMap[view]
	if ok {
		return false
	}
	termsMap[view] = prepreparePayload
	return true
}

func NewInMemoryPBFTStorage() *inMemoryPbftStorage {
	return &inMemoryPbftStorage{
		preprepareStorage: make(map[uint64]map[uint64]*networkcommunication.PrepreparePayload),
	}
}

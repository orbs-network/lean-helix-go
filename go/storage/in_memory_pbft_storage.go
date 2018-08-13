package storage

import (
	"github.com/orbs-network/lean-helix-go/go/networkcommunication"
)

type inMemoryPbftStorage struct {
	preprepareStorage map[uint64]map[uint64]*networkcommunication.PrepreparePayload
	prepareStorage    map[uint64]map[uint64]*networkcommunication.PreparePayload
	commitStorage     map[uint64]map[uint64]*networkcommunication.CommitPayload
	viewChangeStorage map[uint64]map[uint64]*networkcommunication.ViewChangePayload
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
		prepareStorage:    make(map[uint64]map[uint64]*networkcommunication.PreparePayload),
		commitStorage:     make(map[uint64]map[uint64]*networkcommunication.CommitPayload),
		viewChangeStorage: make(map[uint64]map[uint64]*networkcommunication.ViewChangePayload),
	}
}

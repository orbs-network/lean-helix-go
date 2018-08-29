package storage

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/utils"
)

type inMemoryPbftStorage struct {
	// TODO Refactor this mess - in the least create some intermediate types
	preprepareStorage map[lh.BlockHeight]map[lh.ViewCounter]*lh.PrePrepareMessage
	prepareStorage    map[lh.BlockHeight]map[lh.ViewCounter]map[lh.BlockHash]map[lh.PublicKey]*lh.PrepareMessage
	commitStorage     map[lh.BlockHeight]map[lh.ViewCounter]map[lh.BlockHash]map[lh.PublicKey]*lh.CommitMessage
	viewChangeStorage map[lh.BlockHeight]map[lh.ViewCounter]map[lh.PublicKey]*lh.ViewChangeMessage
}

func (storage *inMemoryPbftStorage) StorePrePrepare(ppm *lh.PrePrepareMessage) bool {

	term := ppm.Term
	view := ppm.View

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

func (storage *inMemoryPbftStorage) StorePrepare(pp *lh.PrepareMessage) bool {
	term := pp.Term
	view := pp.View
	// pps -> views ->
	views, ok := storage.prepareStorage[term]
	if !ok {
		views = make(map[lh.ViewCounter]map[lh.BlockHash]map[lh.PublicKey]*lh.PrepareMessage)
		storage.prepareStorage[term] = views
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[lh.BlockHash]map[lh.PublicKey]*lh.PrepareMessage)
		views[view] = blockHashes
	}
	senders, ok := blockHashes[pp.BlockHash]
	if !ok {
		senders = make(map[lh.PublicKey]*lh.PrepareMessage)
		blockHashes[pp.BlockHash] = senders
	}
	senderPublicKey := pp.SignaturePair.SignerPublicKey
	_, ok = senders[senderPublicKey]
	if ok {
		return false
	}
	senders[senderPublicKey] = pp

	utils.Logger.Info("StorePrepare: term=%d view=%d, senderPk=%s, blockHash=%s", term, view, senderPublicKey, pp.BlockHash)

	return true
}

func (storage *inMemoryPbftStorage) StoreCommit(cm *lh.CommitMessage) bool {
	term := cm.Term
	view := cm.View
	// pps -> views ->
	views, ok := storage.commitStorage[term]
	if !ok {
		views = make(map[lh.ViewCounter]map[lh.BlockHash]map[lh.PublicKey]*lh.CommitMessage)
		storage.commitStorage[term] = views
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[lh.BlockHash]map[lh.PublicKey]*lh.CommitMessage)
		views[view] = blockHashes
	}
	senders, ok := blockHashes[cm.BlockHash]
	if !ok {
		senders = make(map[lh.PublicKey]*lh.CommitMessage)
		blockHashes[cm.BlockHash] = senders
	}
	_, ok = senders[cm.SignaturePair.SignerPublicKey]
	if ok {
		return false
	}
	senders[cm.SignaturePair.SignerPublicKey] = cm

	utils.Logger.Info("StoreCommit: term=%d view=%d, senderPk=%s, blockHash=%s", term, view, cm.SignaturePair.SignerPublicKey, cm.BlockHash)

	return true

}

func (storage *inMemoryPbftStorage) StoreViewChange(vcm *lh.ViewChangeMessage) bool {
	term, view := vcm.Term, vcm.View
	// pps -> views ->
	views, ok := storage.viewChangeStorage[term]
	if !ok {
		views = make(map[lh.ViewCounter]map[lh.PublicKey]*lh.ViewChangeMessage)
		storage.viewChangeStorage[term] = views
	}
	senders, ok := views[view]
	if !ok {
		senders = make(map[lh.PublicKey]*lh.ViewChangeMessage)
		views[view] = senders
	}
	_, ok = senders[vcm.SignerPublicKey]
	if ok {
		return false
	}
	senders[vcm.SignerPublicKey] = vcm

	utils.Logger.Info("StoreViewChange: term=%d view=%d, senderPk=%s",
		term, view, vcm.SignerPublicKey)

	return true

}

func (storage *inMemoryPbftStorage) getPrepare(term lh.BlockHeight, view lh.ViewCounter, blockHash lh.BlockHash) (map[lh.PublicKey]*lh.PrepareMessage, bool) {
	views, ok := storage.prepareStorage[term]
	if !ok {
		return nil, false
	}
	blockHashes, ok := views[view]
	if !ok {
		return nil, false
	}
	return blockHashes[blockHash], true
}

func (storage *inMemoryPbftStorage) GetPrepareSendersPKs(term lh.BlockHeight, view lh.ViewCounter, blockHash lh.BlockHash) []lh.PublicKey {
	senders, ok := storage.getPrepare(term, view, blockHash)
	if !ok {
		return []lh.PublicKey{}
	}
	keys := make([]lh.PublicKey, len(senders))
	i := 0
	for k := range senders {
		keys[i] = k
		i++
	}
	return keys
}

func (storage *inMemoryPbftStorage) getCommit(term lh.BlockHeight, view lh.ViewCounter, blockHash lh.BlockHash) (map[lh.PublicKey]*lh.CommitMessage, bool) {
	views, ok := storage.commitStorage[term]
	if !ok {
		return nil, false
	}
	blockHashes, ok := views[view]
	if !ok {
		return nil, false
	}
	return blockHashes[blockHash], true
}

func (storage *inMemoryPbftStorage) GetCommitSendersPKs(term lh.BlockHeight, view lh.ViewCounter, blockHash lh.BlockHash) []lh.PublicKey {
	senders, ok := storage.getCommit(term, view, blockHash)
	if !ok {
		return []lh.PublicKey{}
	}
	keys := make([]lh.PublicKey, len(senders))
	i := 0
	for k := range senders {
		keys[i] = k
		i++
	}
	return keys
}
func (storage *inMemoryPbftStorage) GetViewChangeMessages(term lh.BlockHeight, view lh.ViewCounter, f int) []*lh.ViewChangeMessage {
	views, ok := storage.viewChangeStorage[term]
	if !ok {
		return nil
	}
	senders, ok := views[view]
	if !ok {
		return nil
	}
	minimumNodes := f*2 + 1
	if len(senders) < minimumNodes {
		return nil
	}

	result := make([]*lh.ViewChangeMessage, minimumNodes)
	i := 0
	for _, value := range senders {
		if i >= minimumNodes {
			break
		}
		result[i] = value
		i++
	}
	return result
}

func NewInMemoryPBFTStorage() *inMemoryPbftStorage {
	return &inMemoryPbftStorage{
		preprepareStorage: make(map[lh.BlockHeight]map[lh.ViewCounter]*lh.PrePrepareMessage),
		prepareStorage:    make(map[lh.BlockHeight]map[lh.ViewCounter]map[lh.BlockHash]map[lh.PublicKey]*lh.PrepareMessage),
		commitStorage:     make(map[lh.BlockHeight]map[lh.ViewCounter]map[lh.BlockHash]map[lh.PublicKey]*lh.CommitMessage),
		viewChangeStorage: make(map[lh.BlockHeight]map[lh.ViewCounter]map[lh.PublicKey]*lh.ViewChangeMessage),
	}
}

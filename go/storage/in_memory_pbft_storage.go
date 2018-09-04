package storage

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/utils"
	"sort"
)

type inMemoryPbftStorage struct {
	// TODO Refactor this mess - in the least create some intermediate types
	preprepareStorage map[lh.BlockHeight]map[lh.ViewCounter]*lh.PrePrepareMessage
	prepareStorage    map[lh.BlockHeight]map[lh.ViewCounter]map[lh.BlockHash]map[lh.PublicKey]*lh.PrepareMessage
	commitStorage     map[lh.BlockHeight]map[lh.ViewCounter]map[lh.BlockHash]map[lh.PublicKey]*lh.CommitMessage
	viewChangeStorage map[lh.BlockHeight]map[lh.ViewCounter]map[lh.PublicKey]*lh.ViewChangeMessage
}

func (storage *inMemoryPbftStorage) StorePrePrepare(ppm *lh.PrePrepareMessage) bool {

	term := ppm.Content.Term
	view := ppm.Content.View

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
	term := pp.Content.Term
	view := pp.Content.View
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
	senders, ok := blockHashes[pp.Content.BlockHash]
	if !ok {
		senders = make(map[lh.PublicKey]*lh.PrepareMessage)
		blockHashes[pp.Content.BlockHash] = senders
	}
	senderPublicKey := pp.SignaturePair.SignerPublicKey
	_, ok = senders[senderPublicKey]
	if ok {
		return false
	}
	senders[senderPublicKey] = pp

	utils.Logger.Info("StorePrepare: term=%d view=%d, senderPk=%s, blockHash=%s", term, view, senderPublicKey, pp.Content.BlockHash)

	return true
}

func (storage *inMemoryPbftStorage) StoreCommit(cm *lh.CommitMessage) bool {
	term := cm.Content.Term
	view := cm.Content.View
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
	senders, ok := blockHashes[cm.Content.BlockHash]
	if !ok {
		senders = make(map[lh.PublicKey]*lh.CommitMessage)
		blockHashes[cm.Content.BlockHash] = senders
	}
	_, ok = senders[cm.SignaturePair.SignerPublicKey]
	if ok {
		return false
	}
	senders[cm.SignaturePair.SignerPublicKey] = cm

	utils.Logger.Info("StoreCommit: term=%d view=%d, senderPk=%s, blockHash=%s", term, view, cm.SignaturePair.SignerPublicKey, cm.Content.BlockHash)

	return true

}

func (storage *inMemoryPbftStorage) StoreViewChange(vcm *lh.ViewChangeMessage) bool {
	term, view := vcm.Content.Term, vcm.Content.View
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
	_, ok = senders[vcm.SignaturePair.SignerPublicKey]
	if ok {
		return false
	}
	senders[vcm.SignaturePair.SignerPublicKey] = vcm

	utils.Logger.Info("StoreViewChange: term=%d view=%d, senderPk=%s",
		term, view, vcm.SignaturePair.SignerPublicKey)

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

func (storage *inMemoryPbftStorage) GetLatestPrepared(term lh.BlockHeight, f int) (*lh.PreparedMessages, bool) {
	terms, ok := storage.preprepareStorage[term]
	if !ok {
		return nil, false
	}
	if len(terms) == 0 {
		return nil, false
	}
	views := make([]lh.ViewCounter, 0, len(terms))
	for key, _ := range terms {
		views = append(views, key)
	}
	sort.Sort(lh.ViewCounters(views))
	lastView := views[len(views)-1]

	ppm, ok := storage.getPrePrepareMessage(term, lastView)
	if !ok {
		return nil, false
	}
	prepareMessages, ok := storage.getPrepareMessages(term, lastView, &ppm.Content.BlockHash)
	if len(prepareMessages) < f*2 {
		return nil, false
	}
	return &lh.PreparedMessages{
		PreprepareMessage: ppm,
		PrepareMessages:   prepareMessages,
	}, true

}

func (storage *inMemoryPbftStorage) getPrePrepareMessage(term lh.BlockHeight, view lh.ViewCounter) (*lh.PrePrepareMessage, bool) {
	views, ok := storage.preprepareStorage[term]
	if !ok {
		return nil, false
	}
	result, ok := views[view]
	return result, ok
}

// TODO Whether to use ptr for string (BlockHash)
func (storage *inMemoryPbftStorage) getPrepareMessages(term lh.BlockHeight, view lh.ViewCounter, blockHash *lh.BlockHash) ([]*lh.PrepareMessage, bool) {
	senders, ok := storage.getPrepare(term, view, *blockHash)
	if !ok {
		return nil, false
	}
	values := make([]*lh.PrepareMessage, 0, len(senders))
	for _, v := range senders {
		values = append(values, v)
	}
	return values, true
}

func NewInMemoryPBFTStorage() *inMemoryPbftStorage {
	return &inMemoryPbftStorage{
		preprepareStorage: make(map[lh.BlockHeight]map[lh.ViewCounter]*lh.PrePrepareMessage),
		prepareStorage:    make(map[lh.BlockHeight]map[lh.ViewCounter]map[lh.BlockHash]map[lh.PublicKey]*lh.PrepareMessage),
		commitStorage:     make(map[lh.BlockHeight]map[lh.ViewCounter]map[lh.BlockHash]map[lh.PublicKey]*lh.CommitMessage),
		viewChangeStorage: make(map[lh.BlockHeight]map[lh.ViewCounter]map[lh.PublicKey]*lh.ViewChangeMessage),
	}
}

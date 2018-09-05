package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/types"
	"sort"
)

type inMemoryPbftStorage struct {
	// TODO Refactor this mess - in the least create some intermediate types
	preprepareStorage map[types.BlockHeight]map[types.ViewCounter]*PrePrepareMessage
	prepareStorage    map[types.BlockHeight]map[types.ViewCounter]map[types.BlockHash]map[types.PublicKey]*PrepareMessage
	commitStorage     map[types.BlockHeight]map[types.ViewCounter]map[types.BlockHash]map[types.PublicKey]*CommitMessage
	viewChangeStorage map[types.BlockHeight]map[types.ViewCounter]map[types.PublicKey]*ViewChangeMessage
}

func (storage *inMemoryPbftStorage) StorePrePrepare(ppm *PrePrepareMessage) bool {

	term := ppm.Content.Term
	view := ppm.Content.View

	views, ok := storage.preprepareStorage[term]
	if !ok {
		views = make(map[types.ViewCounter]*PrePrepareMessage)
		storage.preprepareStorage[term] = views
	}

	_, ok = views[view]
	if ok {
		return false
	}
	views[view] = ppm
	return true
}

func (storage *inMemoryPbftStorage) StorePrepare(pp *PrepareMessage) bool {
	term := pp.Content.Term
	view := pp.Content.View
	// pps -> views ->
	views, ok := storage.prepareStorage[term]
	if !ok {
		views = make(map[types.ViewCounter]map[types.BlockHash]map[types.PublicKey]*PrepareMessage)
		storage.prepareStorage[term] = views
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[types.BlockHash]map[types.PublicKey]*PrepareMessage)
		views[view] = blockHashes
	}
	senders, ok := blockHashes[pp.Content.BlockHash]
	if !ok {
		senders = make(map[types.PublicKey]*PrepareMessage)
		blockHashes[pp.Content.BlockHash] = senders
	}
	senderPublicKey := pp.SignaturePair.SignerPublicKey
	_, ok = senders[senderPublicKey]
	if ok {
		return false
	}
	senders[senderPublicKey] = pp

	return true
}

func (storage *inMemoryPbftStorage) StoreCommit(cm *CommitMessage) bool {
	term := cm.Content.Term
	view := cm.Content.View
	// pps -> views ->
	views, ok := storage.commitStorage[term]
	if !ok {
		views = make(map[types.ViewCounter]map[types.BlockHash]map[types.PublicKey]*CommitMessage)
		storage.commitStorage[term] = views
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[types.BlockHash]map[types.PublicKey]*CommitMessage)
		views[view] = blockHashes
	}
	senders, ok := blockHashes[cm.Content.BlockHash]
	if !ok {
		senders = make(map[types.PublicKey]*CommitMessage)
		blockHashes[cm.Content.BlockHash] = senders
	}
	_, ok = senders[cm.SignaturePair.SignerPublicKey]
	if ok {
		return false
	}
	senders[cm.SignaturePair.SignerPublicKey] = cm

	return true

}

func (storage *inMemoryPbftStorage) StoreViewChange(vcm *ViewChangeMessage) bool {
	term, view := vcm.Content.Term, vcm.Content.View
	// pps -> views ->
	views, ok := storage.viewChangeStorage[term]
	if !ok {
		views = make(map[types.ViewCounter]map[types.PublicKey]*ViewChangeMessage)
		storage.viewChangeStorage[term] = views
	}
	senders, ok := views[view]
	if !ok {
		senders = make(map[types.PublicKey]*ViewChangeMessage)
		views[view] = senders
	}
	_, ok = senders[vcm.SignaturePair.SignerPublicKey]
	if ok {
		return false
	}
	senders[vcm.SignaturePair.SignerPublicKey] = vcm

	return true

}

func (storage *inMemoryPbftStorage) getPrepare(term types.BlockHeight, view types.ViewCounter, blockHash types.BlockHash) (map[types.PublicKey]*PrepareMessage, bool) {
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

func (storage *inMemoryPbftStorage) GetPrepareSendersPKs(term types.BlockHeight, view types.ViewCounter, blockHash types.BlockHash) []types.PublicKey {
	senders, ok := storage.getPrepare(term, view, blockHash)
	if !ok {
		return []types.PublicKey{}
	}
	keys := make([]types.PublicKey, len(senders))
	i := 0
	for k := range senders {
		keys[i] = k
		i++
	}
	return keys
}

func (storage *inMemoryPbftStorage) getCommit(term types.BlockHeight, view types.ViewCounter, blockHash types.BlockHash) (map[types.PublicKey]*CommitMessage, bool) {
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

func (storage *inMemoryPbftStorage) GetCommitSendersPKs(term types.BlockHeight, view types.ViewCounter, blockHash types.BlockHash) []types.PublicKey {
	senders, ok := storage.getCommit(term, view, blockHash)
	if !ok {
		return []types.PublicKey{}
	}
	keys := make([]types.PublicKey, len(senders))
	i := 0
	for k := range senders {
		keys[i] = k
		i++
	}
	return keys
}
func (storage *inMemoryPbftStorage) GetViewChangeMessages(term types.BlockHeight, view types.ViewCounter, f int) []*ViewChangeMessage {
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

	result := make([]*ViewChangeMessage, minimumNodes)
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

func (storage *inMemoryPbftStorage) GetLatestPrepared(term types.BlockHeight, f int) (*PreparedMessages, bool) {
	terms, ok := storage.preprepareStorage[term]
	if !ok {
		return nil, false
	}
	if len(terms) == 0 {
		return nil, false
	}
	views := make([]types.ViewCounter, 0, len(terms))
	for key, _ := range terms {
		views = append(views, key)
	}
	sort.Sort(types.ViewCounters(views))
	lastView := views[len(views)-1]

	ppm, ok := storage.getPrePrepareMessage(term, lastView)
	if !ok {
		return nil, false
	}
	prepareMessages, ok := storage.getPrepareMessages(term, lastView, &ppm.Content.BlockHash)
	if len(prepareMessages) < f*2 {
		return nil, false
	}
	return &PreparedMessages{
		PreprepareMessage: ppm,
		PrepareMessages:   prepareMessages,
	}, true

}

func (storage *inMemoryPbftStorage) getPrePrepareMessage(term types.BlockHeight, view types.ViewCounter) (*PrePrepareMessage, bool) {
	views, ok := storage.preprepareStorage[term]
	if !ok {
		return nil, false
	}
	result, ok := views[view]
	return result, ok
}

// TODO Whether to use ptr for string (types.BlockHash)
func (storage *inMemoryPbftStorage) getPrepareMessages(term types.BlockHeight, view types.ViewCounter, blockHash *types.BlockHash) ([]*PrepareMessage, bool) {
	senders, ok := storage.getPrepare(term, view, *blockHash)
	if !ok {
		return nil, false
	}
	values := make([]*PrepareMessage, 0, len(senders))
	for _, v := range senders {
		values = append(values, v)
	}
	return values, true
}

func NewInMemoryPBFTStorage() *inMemoryPbftStorage {
	return &inMemoryPbftStorage{
		preprepareStorage: make(map[types.BlockHeight]map[types.ViewCounter]*PrePrepareMessage),
		prepareStorage:    make(map[types.BlockHeight]map[types.ViewCounter]map[types.BlockHash]map[types.PublicKey]*PrepareMessage),
		commitStorage:     make(map[types.BlockHeight]map[types.ViewCounter]map[types.BlockHash]map[types.PublicKey]*CommitMessage),
		viewChangeStorage: make(map[types.BlockHeight]map[types.ViewCounter]map[types.PublicKey]*ViewChangeMessage),
	}
}

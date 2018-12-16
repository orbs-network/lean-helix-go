package leanhelix

import (
	. "github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"sort"
	"sync"
)

type BlockHashStr string
type PublicKeyStr string

type InMemoryStorage struct {
	mutext            sync.RWMutex
	preprepareStorage map[BlockHeight]map[View]*PreprepareMessage
	prepareStorage    map[BlockHeight]map[View]map[BlockHashStr]map[PublicKeyStr]*PrepareMessage
	commitStorage     map[BlockHeight]map[View]map[BlockHashStr]map[PublicKeyStr]*CommitMessage
	viewChangeStorage map[BlockHeight]map[View]map[PublicKeyStr]*ViewChangeMessage
}

// Preprepare
func (storage *InMemoryStorage) StorePreprepare(ppm *PreprepareMessage) bool {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	height := ppm.BlockHeight()
	view := ppm.Content().SignedHeader().View()

	views, ok := storage.preprepareStorage[height]
	if !ok {
		views = storage.resetPreprepareStorage(height)
	}

	_, ok = views[view]
	if ok {
		return false
	}
	views[view] = ppm
	return true
}

func (storage *InMemoryStorage) GetPreprepareMessage(blockHeight BlockHeight, view View) (*PreprepareMessage, bool) {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	views, ok := storage.preprepareStorage[blockHeight]
	if !ok {
		return nil, false
	}
	result, ok := views[view]
	return result, ok
}

func (storage *InMemoryStorage) GetPreprepareBlock(blockHeight BlockHeight, view View) (Block, bool) {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	views, ok := storage.preprepareStorage[blockHeight]
	if !ok {
		return nil, false
	}
	result, ok := views[view]
	return result.Block(), ok
}

func (storage *InMemoryStorage) GetLatestPreprepare(blockHeight BlockHeight) (*PreprepareMessage, bool) {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	views, ok := storage.preprepareStorage[blockHeight]
	if !ok {
		return nil, false
	}
	if len(views) == 0 {
		return nil, false
	}
	viewKeys := make([]View, 0, len(views))
	for key := range views {
		viewKeys = append(viewKeys, key)
	}
	if len(viewKeys) == 0 {
		return nil, false
	}
	sort.Sort(ViewCounters(viewKeys))
	lastView := viewKeys[len(viewKeys)-1]
	return views[lastView], true
}

// Prepare
func (storage *InMemoryStorage) StorePrepare(pp *PrepareMessage) bool {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	height := pp.BlockHeight()
	view := pp.Content().SignedHeader().View()
	// pps -> views ->
	views, ok := storage.prepareStorage[height]
	if !ok {
		views = storage.resetPrepareStorage(height)
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[BlockHashStr]map[PublicKeyStr]*PrepareMessage)
		views[view] = blockHashes
	}
	ppBlockHash := BlockHashStr(pp.Content().SignedHeader().BlockHash())
	senders, ok := blockHashes[ppBlockHash]
	if !ok {
		senders = make(map[PublicKeyStr]*PrepareMessage)
		blockHashes[ppBlockHash] = senders
	}
	pk := PublicKeyStr(pp.Content().Sender().MemberId())
	_, ok = senders[pk]
	if ok {
		return false
	}
	senders[pk] = pp

	return true
}

func (storage *InMemoryStorage) getPrepare(blockHeight BlockHeight, view View, blockHash BlockHash) (map[PublicKeyStr]*PrepareMessage, bool) {
	views, ok := storage.prepareStorage[blockHeight]
	if !ok {
		return nil, false
	}
	blockHashes, ok := views[view]
	if !ok {
		return nil, false
	}
	return blockHashes[BlockHashStr(blockHash)], true
}

func (storage *InMemoryStorage) GetPrepareMessages(blockHeight BlockHeight, view View, blockHash BlockHash) ([]*PrepareMessage, bool) {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	senders, ok := storage.getPrepare(blockHeight, view, blockHash)
	if !ok {
		return nil, false
	}
	values := make([]*PrepareMessage, 0, len(senders))
	for _, v := range senders {
		values = append(values, v)
	}
	return values, true
}

func (storage *InMemoryStorage) GetPrepareSendersPKs(blockHeight BlockHeight, view View, blockHash BlockHash) []MemberId {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	senders, ok := storage.getPrepare(blockHeight, view, blockHash)
	if !ok {
		return []MemberId{}
	}
	keys := make([]MemberId, len(senders))
	i := 0
	for k := range senders {
		keys[i] = MemberId(k)
		i++
	}
	return keys
}

// Commit
func (storage *InMemoryStorage) StoreCommit(cm *CommitMessage) bool {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	height := cm.Content().SignedHeader().BlockHeight()
	view := cm.Content().SignedHeader().View()
	// pps -> views ->
	views, ok := storage.commitStorage[height]
	if !ok {
		views = storage.resetCommitStorage(height)
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[BlockHashStr]map[PublicKeyStr]*CommitMessage)
		views[view] = blockHashes
	}
	cmBlockHash := BlockHashStr(cm.Content().SignedHeader().BlockHash())
	senders, ok := blockHashes[cmBlockHash]
	if !ok {
		senders = make(map[PublicKeyStr]*CommitMessage)
		blockHashes[cmBlockHash] = senders
	}
	pk := PublicKeyStr(cm.Content().Sender().MemberId())
	_, ok = senders[pk]
	if ok {
		return false
	}
	senders[pk] = cm

	return true

}

func (storage *InMemoryStorage) getCommit(blockHeight BlockHeight, view View, blockHash BlockHash) (map[PublicKeyStr]*CommitMessage, bool) {
	views, ok := storage.commitStorage[blockHeight]
	if !ok {
		return nil, false
	}
	blockHashes, ok := views[view]
	if !ok {
		return nil, false
	}
	return blockHashes[BlockHashStr(blockHash)], true
}

func (storage *InMemoryStorage) GetCommitMessages(blockHeight BlockHeight, view View, blockHash BlockHash) ([]*CommitMessage, bool) {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	senders, ok := storage.getCommit(blockHeight, view, blockHash)
	if !ok {
		return nil, false
	}
	values := make([]*CommitMessage, 0, len(senders))
	for _, v := range senders {
		values = append(values, v)
	}
	return values, true
}

func (storage *InMemoryStorage) GetCommitSendersPKs(blockHeight BlockHeight, view View, blockHash BlockHash) []MemberId {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	senders, ok := storage.getCommit(blockHeight, view, blockHash)
	if !ok {
		return []MemberId{}
	}
	keys := make([]MemberId, len(senders))
	i := 0
	for k := range senders {
		keys[i] = MemberId(k)
		i++
	}
	return keys
}

// View Change
func (storage *InMemoryStorage) StoreViewChange(vcm *ViewChangeMessage) bool {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	height, view := vcm.Content().SignedHeader().BlockHeight(), vcm.Content().SignedHeader().View()
	// pps -> views ->
	views, ok := storage.viewChangeStorage[height]
	if !ok {
		views = storage.resetViewChangeStorage(height)
	}
	senders, ok := views[view]
	if !ok {
		senders = make(map[PublicKeyStr]*ViewChangeMessage)
		views[view] = senders
	}

	pk := PublicKeyStr(vcm.Content().Sender().MemberId())
	_, ok = senders[pk]
	if ok {
		return false
	}
	senders[pk] = vcm

	return true

}

func (storage *InMemoryStorage) GetViewChangeMessages(blockHeight BlockHeight, view View) ([]*ViewChangeMessage, bool) {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	views, ok := storage.viewChangeStorage[blockHeight]
	if !ok {
		return nil, false
	}
	senders, ok := views[view]
	if !ok {
		return nil, false
	}

	result := make([]*ViewChangeMessage, len(senders))
	i := 0
	for _, value := range senders {
		result[i] = value
		i++
	}
	return result, true
}

func (storage *InMemoryStorage) ClearBlockHeightLogs(blockHeight BlockHeight) {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	storage.resetPreprepareStorage(blockHeight)
	storage.resetPrepareStorage(blockHeight)
	storage.resetCommitStorage(blockHeight)
	storage.resetViewChangeStorage(blockHeight)
}

func (storage *InMemoryStorage) resetPreprepareStorage(blockHeight BlockHeight) map[View]*PreprepareMessage {
	views := make(map[View]*PreprepareMessage) // map[View]
	storage.preprepareStorage[blockHeight] = views
	return views
}

func (storage *InMemoryStorage) resetPrepareStorage(blockHeight BlockHeight) map[View]map[BlockHashStr]map[PublicKeyStr]*PrepareMessage {
	views := make(map[View]map[BlockHashStr]map[PublicKeyStr]*PrepareMessage)
	storage.prepareStorage[blockHeight] = views
	return views
}

func (storage *InMemoryStorage) resetCommitStorage(blockHeight BlockHeight) map[View]map[BlockHashStr]map[PublicKeyStr]*CommitMessage {
	views := make(map[View]map[BlockHashStr]map[PublicKeyStr]*CommitMessage)
	storage.commitStorage[blockHeight] = views
	return views
}
func (storage *InMemoryStorage) resetViewChangeStorage(blockHeight BlockHeight) map[View]map[PublicKeyStr]*ViewChangeMessage {
	views := make(map[View]map[PublicKeyStr]*ViewChangeMessage)
	storage.viewChangeStorage[blockHeight] = views
	return views
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		preprepareStorage: make(map[BlockHeight]map[View]*PreprepareMessage),
		prepareStorage:    make(map[BlockHeight]map[View]map[BlockHashStr]map[PublicKeyStr]*PrepareMessage),
		commitStorage:     make(map[BlockHeight]map[View]map[BlockHashStr]map[PublicKeyStr]*CommitMessage),
		viewChangeStorage: make(map[BlockHeight]map[View]map[PublicKeyStr]*ViewChangeMessage),
	}
}

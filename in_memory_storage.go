package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"sort"
	"sync"
)

type BlockHashStr string
type MemberIdStr string

// Sorting View arrays
type viewCounters []primitives.View

func (arr viewCounters) Len() int           { return len(arr) }
func (arr viewCounters) Swap(i, j int)      { arr[i], arr[j] = arr[j], arr[i] }
func (arr viewCounters) Less(i, j int) bool { return arr[i] < arr[j] }

type InMemoryStorage struct {
	mutext            sync.RWMutex
	preprepareStorage map[primitives.BlockHeight]map[primitives.View]*PreprepareMessage
	prepareStorage    map[primitives.BlockHeight]map[primitives.View]map[BlockHashStr]map[MemberIdStr]*PrepareMessage
	commitStorage     map[primitives.BlockHeight]map[primitives.View]map[BlockHashStr]map[MemberIdStr]*CommitMessage
	viewChangeStorage map[primitives.BlockHeight]map[primitives.View]map[MemberIdStr]*ViewChangeMessage
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

func (storage *InMemoryStorage) GetPreprepareMessage(blockHeight primitives.BlockHeight, view primitives.View) (*PreprepareMessage, bool) {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	views, ok := storage.preprepareStorage[blockHeight]
	if !ok {
		return nil, false
	}
	result, ok := views[view]
	return result, ok
}

func (storage *InMemoryStorage) GetPreprepareBlock(blockHeight primitives.BlockHeight, view primitives.View) (Block, bool) {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	views, ok := storage.preprepareStorage[blockHeight]
	if !ok {
		return nil, false
	}
	result, ok := views[view]
	return result.Block(), ok
}

func (storage *InMemoryStorage) GetLatestPreprepare(blockHeight primitives.BlockHeight) (*PreprepareMessage, bool) {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	views, ok := storage.preprepareStorage[blockHeight]
	if !ok {
		return nil, false
	}
	if len(views) == 0 {
		return nil, false
	}
	viewKeys := make([]primitives.View, 0, len(views))
	for key := range views {
		viewKeys = append(viewKeys, key)
	}
	if len(viewKeys) == 0 {
		return nil, false
	}
	sort.Sort(viewCounters(viewKeys))
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
		blockHashes = make(map[BlockHashStr]map[MemberIdStr]*PrepareMessage)
		views[view] = blockHashes
	}
	ppBlockHash := BlockHashStr(pp.Content().SignedHeader().BlockHash())
	senders, ok := blockHashes[ppBlockHash]
	if !ok {
		senders = make(map[MemberIdStr]*PrepareMessage)
		blockHashes[ppBlockHash] = senders
	}
	id := MemberIdStr(pp.Content().Sender().MemberId())
	_, ok = senders[id]
	if ok {
		return false
	}
	senders[id] = pp

	return true
}

func (storage *InMemoryStorage) getPrepare(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) (map[MemberIdStr]*PrepareMessage, bool) {
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

func (storage *InMemoryStorage) GetPrepareMessages(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) ([]*PrepareMessage, bool) {
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

func (storage *InMemoryStorage) GetPrepareSendersIds(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) []primitives.MemberId {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	senders, ok := storage.getPrepare(blockHeight, view, blockHash)
	if !ok {
		return []primitives.MemberId{}
	}
	keys := make([]primitives.MemberId, len(senders))
	i := 0
	for k := range senders {
		keys[i] = primitives.MemberId(k)
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
		blockHashes = make(map[BlockHashStr]map[MemberIdStr]*CommitMessage)
		views[view] = blockHashes
	}
	cmBlockHash := BlockHashStr(cm.Content().SignedHeader().BlockHash())
	senders, ok := blockHashes[cmBlockHash]
	if !ok {
		senders = make(map[MemberIdStr]*CommitMessage)
		blockHashes[cmBlockHash] = senders
	}
	id := MemberIdStr(cm.Content().Sender().MemberId())
	_, ok = senders[id]
	if ok {
		return false
	}
	senders[id] = cm

	return true

}

func (storage *InMemoryStorage) getCommit(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) (map[MemberIdStr]*CommitMessage, bool) {
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

func (storage *InMemoryStorage) GetCommitMessages(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) ([]*CommitMessage, bool) {
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

func (storage *InMemoryStorage) GetCommitSendersIds(blockHeight primitives.BlockHeight, view primitives.View, blockHash primitives.BlockHash) []primitives.MemberId {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	senders, ok := storage.getCommit(blockHeight, view, blockHash)
	if !ok {
		return []primitives.MemberId{}
	}
	keys := make([]primitives.MemberId, len(senders))
	i := 0
	for k := range senders {
		keys[i] = primitives.MemberId(k)
		i++
	}
	return keys
}

// primitives.View Change
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
		senders = make(map[MemberIdStr]*ViewChangeMessage)
		views[view] = senders
	}

	id := MemberIdStr(vcm.Content().Sender().MemberId())
	_, ok = senders[id]
	if ok {
		return false
	}
	senders[id] = vcm

	return true

}

func (storage *InMemoryStorage) GetViewChangeMessages(blockHeight primitives.BlockHeight, view primitives.View) ([]*ViewChangeMessage, bool) {
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

func (storage *InMemoryStorage) ClearBlockHeightLogs(blockHeight primitives.BlockHeight) {
	storage.mutext.Lock()
	defer storage.mutext.Unlock()

	storage.resetPreprepareStorage(blockHeight)
	storage.resetPrepareStorage(blockHeight)
	storage.resetCommitStorage(blockHeight)
	storage.resetViewChangeStorage(blockHeight)
}

func (storage *InMemoryStorage) resetPreprepareStorage(blockHeight primitives.BlockHeight) map[primitives.View]*PreprepareMessage {
	views := make(map[primitives.View]*PreprepareMessage) // map[primitives.View]
	storage.preprepareStorage[blockHeight] = views
	return views
}

func (storage *InMemoryStorage) resetPrepareStorage(blockHeight primitives.BlockHeight) map[primitives.View]map[BlockHashStr]map[MemberIdStr]*PrepareMessage {
	views := make(map[primitives.View]map[BlockHashStr]map[MemberIdStr]*PrepareMessage)
	storage.prepareStorage[blockHeight] = views
	return views
}

func (storage *InMemoryStorage) resetCommitStorage(blockHeight primitives.BlockHeight) map[primitives.View]map[BlockHashStr]map[MemberIdStr]*CommitMessage {
	views := make(map[primitives.View]map[BlockHashStr]map[MemberIdStr]*CommitMessage)
	storage.commitStorage[blockHeight] = views
	return views
}
func (storage *InMemoryStorage) resetViewChangeStorage(blockHeight primitives.BlockHeight) map[primitives.View]map[MemberIdStr]*ViewChangeMessage {
	views := make(map[primitives.View]map[MemberIdStr]*ViewChangeMessage)
	storage.viewChangeStorage[blockHeight] = views
	return views
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		preprepareStorage: make(map[primitives.BlockHeight]map[primitives.View]*PreprepareMessage),
		prepareStorage:    make(map[primitives.BlockHeight]map[primitives.View]map[BlockHashStr]map[MemberIdStr]*PrepareMessage),
		commitStorage:     make(map[primitives.BlockHeight]map[primitives.View]map[BlockHashStr]map[MemberIdStr]*CommitMessage),
		viewChangeStorage: make(map[primitives.BlockHeight]map[primitives.View]map[MemberIdStr]*ViewChangeMessage),
	}
}

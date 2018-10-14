package leanhelix

type BlockHashStr string
type PublicKeyStr string

type Storage interface {
	StorePreprepare(ppm PreprepareMessage) bool
	StorePrepare(pp PrepareMessage) bool
	StoreCommit(cm CommitMessage) bool
	StoreViewChange(vcm ViewChangeMessage) bool
	GetPrepareSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519_public_key
	GetCommitSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519_public_key
	GetViewChangeMessages(blockHeight BlockHeight, view View, f int) []ViewChangeMessage
	GetPreprepare(blockHeight BlockHeight, view View) (PreprepareMessage, bool)
	GetPrepares(blockHeight BlockHeight, view View, blockHash Uint256) ([]PrepareMessage, bool)
	GetLatestPrepared(blockHeight BlockHeight, f int) (PreparedProof, bool)
	ClearTermLogs(blockHeight BlockHeight)
}

type InMemoryStorage struct {
	// TODO Refactor this mess - in the least create some intermediate primitives
	preprepareStorage map[BlockHeight]map[View]PreprepareMessage
	prepareStorage    map[BlockHeight]map[View]map[Uint256]map[Ed25519_public_key]PrepareMessage
	commitStorage     map[BlockHeight]map[View]map[Uint256]map[Ed25519_public_key]CommitMessage
	viewChangeStorage map[BlockHeight]map[View]map[Ed25519_public_key]ViewChangeMessage
}

func (storage *InMemoryStorage) StorePreprepare(ppm PreprepareMessage) bool {

	height := ppm.SignedHeader().BlockHeight()
	view := ppm.SignedHeader().View()

	views, ok := storage.preprepareStorage[*height]
	if !ok {
		views = storage.resetPreprepareStorage(*height)
	}

	_, ok = views[*view]
	if ok {
		return false
	}
	views[*view] = ppm
	return true
}

func (storage *InMemoryStorage) StorePrepare(pp PrepareMessage) bool {
	height := *pp.SignedHeader().BlockHeight()
	view := *pp.SignedHeader().View()
	// pps -> views ->
	views, ok := storage.prepareStorage[height]
	if !ok {
		views = storage.resetPrepareStorage(height)
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[Uint256]map[Ed25519_public_key]PrepareMessage)
		views[view] = blockHashes
	}
	ppBlockHash := *pp.SignedHeader().BlockHash()
	senders, ok := blockHashes[ppBlockHash]
	if !ok {
		senders = make(map[Ed25519_public_key]PrepareMessage)
		blockHashes[ppBlockHash] = senders
	}
	pk := *pp.Sender().SenderPublicKey()
	_, ok = senders[pk]
	if ok {
		return false
	}
	senders[pk] = pp

	return true
}

func (storage *InMemoryStorage) StoreCommit(cm CommitMessage) bool {
	height := cm.SignedHeader().BlockHeight()
	view := cm.SignedHeader().View()
	// pps -> views ->
	views, ok := storage.commitStorage[*height]
	if !ok {
		views = storage.resetCommitStorage(*height)
	}

	blockHashes, ok := views[*view]
	if !ok {
		blockHashes = make(map[Uint256]map[Ed25519_public_key]CommitMessage)
		views[*view] = blockHashes
	}
	senders, ok := blockHashes[*cm.SignedHeader().BlockHash()]
	if !ok {
		senders = make(map[Ed25519_public_key]CommitMessage)
		blockHashes[*cm.SignedHeader().BlockHash()] = senders
	}
	pk := *cm.Sender().SenderPublicKey()
	_, ok = senders[pk]
	if ok {
		return false
	}
	senders[pk] = cm

	return true

}

func (storage *InMemoryStorage) StoreViewChange(vcm ViewChangeMessage) bool {
	height, view := vcm.SignedHeader().BlockHeight(), vcm.SignedHeader().View()
	// pps -> views ->
	views, ok := storage.viewChangeStorage[*height]
	if !ok {
		views = storage.resetViewChangeStorage(*height)
	}
	senders, ok := views[*view]
	if !ok {
		senders = make(map[Ed25519_public_key]ViewChangeMessage)
		views[*view] = senders
	}
	pk := *vcm.Sender().SenderPublicKey()
	_, ok = senders[pk]
	if ok {
		return false
	}
	senders[pk] = vcm

	return true

}

func (storage *InMemoryStorage) getPrepare(blockHeight BlockHeight, view View, blockHash Uint256) (map[Ed25519_public_key]PrepareMessage, bool) {
	views, ok := storage.prepareStorage[blockHeight]
	if !ok {
		return nil, false
	}
	blockHashes, ok := views[view]
	if !ok {
		return nil, false
	}
	return blockHashes[blockHash], true
}

func (storage *InMemoryStorage) GetPrepareSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519_public_key {
	senders, ok := storage.getPrepare(blockHeight, view, blockHash)
	if !ok {
		return []Ed25519_public_key{}
	}
	keys := make([]Ed25519_public_key, len(senders))
	i := 0
	for k := range senders {
		keys[i] = k
		i++
	}
	return keys
}

func (storage *InMemoryStorage) getCommit(blockHeight BlockHeight, view View, blockHash Uint256) (map[Ed25519_public_key]CommitMessage, bool) {
	views, ok := storage.commitStorage[blockHeight]
	if !ok {
		return nil, false
	}
	blockHashes, ok := views[view]
	if !ok {
		return nil, false
	}
	return blockHashes[blockHash], true
}

func (storage *InMemoryStorage) GetCommitSendersPKs(blockHeight BlockHeight, view View, blockHash Uint256) []Ed25519_public_key {
	senders, ok := storage.getCommit(blockHeight, view, blockHash)
	if !ok {
		return []Ed25519_public_key{}
	}
	keys := make([]Ed25519_public_key, len(senders))
	i := 0
	for k := range senders {
		keys[i] = Ed25519_public_key(k)
		i++
	}
	return keys
}
func (storage *InMemoryStorage) GetViewChangeMessages(blockHeight BlockHeight, view View, f int) []ViewChangeMessage {
	views, ok := storage.viewChangeStorage[blockHeight]
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

	result := make([]ViewChangeMessage, minimumNodes)
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

func (storage *InMemoryStorage) GetPreprepare(blockHeight BlockHeight, view View) (PreprepareMessage, bool) {
	views, ok := storage.preprepareStorage[blockHeight]
	if !ok {
		return nil, false
	}
	result, ok := views[view]
	return result, ok
}

// TODO Whether to use ptr for string (BlockHash)
func (storage *InMemoryStorage) GetPrepares(blockHeight BlockHeight, view View, blockHash Uint256) ([]PrepareMessage, bool) {
	senders, ok := storage.getPrepare(blockHeight, view, blockHash)
	if !ok {
		return nil, false
	}
	values := make([]PrepareMessage, 0, len(senders))
	for _, v := range senders {
		values = append(values, v)
	}
	return values, true
}

//func (storage *InMemoryStorage) GetLatestPrepared(blockHeight BlockHeight, f int) (PreparedProof, bool) {
//	views, ok := storage.preprepareStorage[blockHeight]
//	if !ok {
//		return nil, false
//	}
//	if len(views) == 0 {
//		return nil, false
//	}
//	viewKeys := make([]View, 0, len(views))
//	for key := range views {
//		viewKeys = append(viewKeys, key)
//	}
//	sort.Sort(ViewCounters(viewKeys))
//	lastView := viewKeys[len(viewKeys)-1]
//
//	lastViewPpm, ok := storage.GetPreprepare(blockHeight, lastView)
//	if !ok {
//		return nil, false
//	}
//	prepareMessages, ok := storage.GetPrepares(blockHeight, lastView, lastViewPpm.SignedHeader().BlockHash())
//	if len(prepareMessages) < f*2 {
//		return nil, false
//	}
//
//	proof := CreatePreparedProof(lastViewPpm, prepareMessages)
//	return proof, true
//
//}

// TODO Keep this name? it means the same as Term in LeanHelixTerm
func (storage *InMemoryStorage) ClearTermLogs(blockHeight BlockHeight) {
	storage.resetPreprepareStorage(blockHeight)
	storage.resetPrepareStorage(blockHeight)
	storage.resetCommitStorage(blockHeight)
	storage.resetViewChangeStorage(blockHeight)
}

func (storage *InMemoryStorage) resetPreprepareStorage(blockHeight BlockHeight) map[View]PreprepareMessage {
	views := make(map[View]PreprepareMessage) // map[View]
	storage.preprepareStorage[blockHeight] = views
	return views
}

func (storage *InMemoryStorage) resetPrepareStorage(blockHeight BlockHeight) map[View]map[Uint256]map[Ed25519_public_key]PrepareMessage {
	views := make(map[View]map[Uint256]map[Ed25519_public_key]PrepareMessage)
	storage.prepareStorage[blockHeight] = views
	return views
}

func (storage *InMemoryStorage) resetCommitStorage(blockHeight BlockHeight) map[View]map[Uint256]map[Ed25519_public_key]CommitMessage {
	views := make(map[View]map[Uint256]map[Ed25519_public_key]CommitMessage)
	storage.commitStorage[blockHeight] = views
	return views
}
func (storage *InMemoryStorage) resetViewChangeStorage(blockHeight BlockHeight) map[View]map[Ed25519_public_key]ViewChangeMessage {
	views := make(map[View]map[Ed25519_public_key]ViewChangeMessage)
	storage.viewChangeStorage[blockHeight] = views
	return views
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		preprepareStorage: make(map[BlockHeight]map[View]PreprepareMessage),
		prepareStorage:    make(map[BlockHeight]map[View]map[Uint256]map[Ed25519_public_key]PrepareMessage),
		commitStorage:     make(map[BlockHeight]map[View]map[Uint256]map[Ed25519_public_key]CommitMessage),
		viewChangeStorage: make(map[BlockHeight]map[View]map[Ed25519_public_key]ViewChangeMessage),
	}
}

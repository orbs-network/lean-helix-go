package leanhelix

import "sort"

type BlockHashStr string
type PublicKeyStr string

type InMemoryPbftStorage struct {
	// TODO Refactor this mess - in the least create some intermediate types
	preprepareStorage map[BlockHeight]map[ViewCounter]PreprepareMessage
	prepareStorage    map[BlockHeight]map[ViewCounter]map[BlockHashStr]map[PublicKeyStr]PrepareMessage
	commitStorage     map[BlockHeight]map[ViewCounter]map[BlockHashStr]map[PublicKeyStr]CommitMessage
	viewChangeStorage map[BlockHeight]map[ViewCounter]map[PublicKeyStr]ViewChangeMessage
}

func (storage *InMemoryPbftStorage) StorePrePrepare(ppm PreprepareMessage) bool {

	term := ppm.Term()
	view := ppm.View()

	views, ok := storage.preprepareStorage[term]
	if !ok {
		views = make(map[ViewCounter]PreprepareMessage)
		storage.preprepareStorage[term] = views
	}

	_, ok = views[view]
	if ok {
		return false
	}
	views[view] = ppm
	return true
}

func (storage *InMemoryPbftStorage) StorePrepare(pp PrepareMessage) bool {
	term := pp.Term()
	view := pp.View()
	// pps -> views ->
	views, ok := storage.prepareStorage[term]
	if !ok {
		views = make(map[ViewCounter]map[BlockHashStr]map[PublicKeyStr]PrepareMessage)
		storage.prepareStorage[term] = views
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[BlockHashStr]map[PublicKeyStr]PrepareMessage)
		views[view] = blockHashes
	}
	ppBlockHash := BlockHashStr(pp.BlockHash())
	senders, ok := blockHashes[ppBlockHash]
	if !ok {
		senders = make(map[PublicKeyStr]PrepareMessage)
		blockHashes[ppBlockHash] = senders
	}
	pk := PublicKeyStr(pp.Sender().SenderPublicKey())
	_, ok = senders[pk]
	if ok {
		return false
	}
	senders[pk] = pp

	return true
}

func (storage *InMemoryPbftStorage) StoreCommit(cm CommitMessage) bool {
	term := cm.Term()
	view := cm.View()
	// pps -> views ->
	views, ok := storage.commitStorage[term]
	if !ok {
		views = make(map[ViewCounter]map[BlockHashStr]map[PublicKeyStr]CommitMessage)
		storage.commitStorage[term] = views
	}

	blockHashes, ok := views[view]
	if !ok {
		blockHashes = make(map[BlockHashStr]map[PublicKeyStr]CommitMessage)
		views[view] = blockHashes
	}
	senders, ok := blockHashes[BlockHashStr(cm.BlockHash())]
	if !ok {
		senders = make(map[PublicKeyStr]CommitMessage)
		blockHashes[BlockHashStr(cm.BlockHash())] = senders
	}
	pk := PublicKeyStr(cm.Sender().SenderPublicKey())
	_, ok = senders[pk]
	if ok {
		return false
	}
	senders[pk] = cm

	return true

}

func (storage *InMemoryPbftStorage) StoreViewChange(vcm ViewChangeMessage) bool {
	term, view := vcm.Term(), vcm.View()
	// pps -> views ->
	views, ok := storage.viewChangeStorage[term]
	if !ok {
		views = make(map[ViewCounter]map[PublicKeyStr]ViewChangeMessage)
		storage.viewChangeStorage[term] = views
	}
	senders, ok := views[view]
	if !ok {
		senders = make(map[PublicKeyStr]ViewChangeMessage)
		views[view] = senders
	}
	pk := PublicKeyStr(vcm.Sender().SenderPublicKey())
	_, ok = senders[pk]
	if ok {
		return false
	}
	senders[pk] = vcm

	return true

}

func (storage *InMemoryPbftStorage) getPrepare(term BlockHeight, view ViewCounter, blockHash BlockHash) (map[PublicKeyStr]PrepareMessage, bool) {
	views, ok := storage.prepareStorage[term]
	if !ok {
		return nil, false
	}
	blockHashes, ok := views[view]
	if !ok {
		return nil, false
	}
	return blockHashes[BlockHashStr(blockHash)], true
}

func (storage *InMemoryPbftStorage) GetPrepareSendersPKs(term BlockHeight, view ViewCounter, blockHash BlockHash) []PublicKey {
	senders, ok := storage.getPrepare(term, view, blockHash)
	if !ok {
		return []PublicKey{}
	}
	keys := make([]PublicKey, len(senders))
	i := 0
	for k := range senders {
		keys[i] = PublicKey(k)
		i++
	}
	return keys
}

func (storage *InMemoryPbftStorage) getCommit(term BlockHeight, view ViewCounter, blockHash BlockHash) (map[PublicKeyStr]CommitMessage, bool) {
	views, ok := storage.commitStorage[term]
	if !ok {
		return nil, false
	}
	blockHashes, ok := views[view]
	if !ok {
		return nil, false
	}
	return blockHashes[BlockHashStr(blockHash)], true
}

func (storage *InMemoryPbftStorage) GetCommitSendersPKs(term BlockHeight, view ViewCounter, blockHash BlockHash) []PublicKey {
	senders, ok := storage.getCommit(term, view, blockHash)
	if !ok {
		return []PublicKey{}
	}
	keys := make([]PublicKey, len(senders))
	i := 0
	for k := range senders {
		keys[i] = PublicKey(k)
		i++
	}
	return keys
}
func (storage *InMemoryPbftStorage) GetViewChangeMessages(term BlockHeight, view ViewCounter, f int) []ViewChangeMessage {
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

func (storage *InMemoryPbftStorage) getPreprepareMessage(term BlockHeight, view ViewCounter) (PreprepareMessage, bool) {
	views, ok := storage.preprepareStorage[term]
	if !ok {
		return nil, false
	}
	result, ok := views[view]
	return result, ok
}

// TODO Whether to use ptr for string (BlockHash)
func (storage *InMemoryPbftStorage) getPrepareMessages(term BlockHeight, view ViewCounter, blockHash BlockHash) ([]PrepareMessage, bool) {
	senders, ok := storage.getPrepare(term, view, blockHash)
	if !ok {
		return nil, false
	}
	values := make([]PrepareMessage, 0, len(senders))
	for _, v := range senders {
		values = append(values, v)
	}
	return values, true
}

func (storage *InMemoryPbftStorage) GetLatestPrepared(term BlockHeight, f int) (PreparedProof, bool) {
	terms, ok := storage.preprepareStorage[term]
	if !ok {
		return nil, false
	}
	if len(terms) == 0 {
		return nil, false
	}
	views := make([]ViewCounter, 0, len(terms))
	for key, _ := range terms {
		views = append(views, key)
	}
	sort.Sort(ViewCounters(views))
	lastView := views[len(views)-1]

	ppm, ok := storage.getPreprepareMessage(term, lastView)
	if !ok {
		return nil, false
	}
	prepareMessages, ok := storage.getPrepareMessages(term, lastView, ppm.BlockHash())
	if len(prepareMessages) < f*2 {
		return nil, false
	}

	proof := CreatePreparedProof(ppm, prepareMessages)
	return proof, true

}

func NewInMemoryPBFTStorage() *InMemoryPbftStorage {
	return &InMemoryPbftStorage{
		preprepareStorage: make(map[BlockHeight]map[ViewCounter]PreprepareMessage),
		prepareStorage:    make(map[BlockHeight]map[ViewCounter]map[BlockHashStr]map[PublicKeyStr]PrepareMessage),
		commitStorage:     make(map[BlockHeight]map[ViewCounter]map[BlockHashStr]map[PublicKeyStr]CommitMessage),
		viewChangeStorage: make(map[BlockHeight]map[ViewCounter]map[PublicKeyStr]ViewChangeMessage),
	}
}

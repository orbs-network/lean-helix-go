package leanhelix

type LeanHelix interface {
	RegisterOnCommitted(cb func(block Block))
	Dispose()
	Start(height BlockHeight)
	IsLeader() bool
}

type leanHelix struct {
}

func NewLeanHelix(config *Config) LeanHelix {
	return &leanHelix{}
}

func (lh *leanHelix) RegisterOnCommitted(cb func(block Block)) {
	// TODO: implement
}

func (lh *leanHelix) Dispose() {
	// TODO: implement
}

func (lh *leanHelix) Start(height BlockHeight) {
	// TODO: implement
}

func (lh *leanHelix) IsLeader() bool {
	// TODO: implement
	return false
}

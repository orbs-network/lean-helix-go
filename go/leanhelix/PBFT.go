package leanhelix

type PBFT struct {
}

func NewPBFT(config *Config) *PBFT {
	return &PBFT{}
}

func (pbft *PBFT) RegisterOnCommitted(cb func(block *Block)) {
	// TODO: implement
}

func (pbft *PBFT) Dispose() {
	// TODO: implement
}

func (pbft *PBFT) Start(height BlockHeight) {
	// TODO: implement
}

func (pbft *PBFT) IsLeader() bool {
	// TODO: implement
	return false
}

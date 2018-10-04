package leanhelix

// External interfaces of this library (temporary)

type BlockHeader interface {
}

// TODO refactor
type PreparedProofInternal struct {
	preprepare PreprepareMessage
	prepares   []PrepareMessage
}

func (pf *PreparedProofInternal) PreprepareMessage() PreprepareMessage {
	return pf.preprepare
}

func (pf *PreparedProofInternal) PrepareMessages() []PrepareMessage {
	return pf.prepares
}

func CreatePreparedProof(ppm PreprepareMessage, pms []PrepareMessage) PreparedProof {
	return &PreparedProofInternal{
		preprepare: ppm,
		prepares:   pms,
	}
}

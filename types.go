package leanhelix

type BlockHeight uint64

func (h BlockHeight) String() string {
	return string(h)
}

func (h BlockHeight) Equals(other BlockHeight) bool {
	return uint64(h) == uint64(other)
}

type View uint64

func (v View) String() string {
	return string(v)
}

func (v View) Equals(other View) bool {
	return uint64(v) == uint64(other)
}

type BlockHash []byte

func (hash BlockHash) String() string {
	return string(hash)
}

func (hash BlockHash) Equals(other BlockHash) bool {
	return string(hash) == string(other)
}

type PublicKey []byte

func (pk PublicKey) String() string {
	return string(pk)
}
func (pk PublicKey) Equals(other PublicKey) bool {
	return string(pk) == string(other)
}

type Signature []byte

func (s Signature) String() string {
	return string(s)
}
func (s Signature) Equals(other Signature) bool {
	return string(s) == string(other)
}

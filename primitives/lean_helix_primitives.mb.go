// AUTO GENERATED FILE (by membufc proto compiler v0.0.20)
package primitives

import (
	"bytes"
	"fmt"
)

type Bls1Sig []byte

func (x Bls1Sig) String() string {
	return fmt.Sprintf("%x", []byte(x))
}

func (x Bls1Sig) Equal(y Bls1Sig) bool {
	return bytes.Equal(x, y)
}

func (x Bls1Sig) KeyForMap() string {
	return string(x)
}

type Ed25519PublicKey []byte

func (x Ed25519PublicKey) String() string {
	return fmt.Sprintf("%x", []byte(x))
}

func (x Ed25519PublicKey) Equal(y Ed25519PublicKey) bool {
	return bytes.Equal(x, y)
}

func (x Ed25519PublicKey) KeyForMap() string {
	return string(x)
}

type Ed25519Sig []byte

func (x Ed25519Sig) String() string {
	return fmt.Sprintf("%x", []byte(x))
}

func (x Ed25519Sig) Equal(y Ed25519Sig) bool {
	return bytes.Equal(x, y)
}

func (x Ed25519Sig) KeyForMap() string {
	return string(x)
}

type Uint256 []byte

func (x Uint256) String() string {
	return fmt.Sprintf("%x", []byte(x))
}

func (x Uint256) Equal(y Uint256) bool {
	return bytes.Equal(x, y)
}

func (x Uint256) KeyForMap() string {
	return string(x)
}

type BlockHeight uint64

func (x BlockHeight) String() string {
	return fmt.Sprintf("%x", uint64(x))
}

func (x BlockHeight) Equal(y BlockHeight) bool {
	return x == y
}

func (x BlockHeight) KeyForMap() uint64 {
	return uint64(x)
}

type View uint64

func (x View) String() string {
	return fmt.Sprintf("%x", uint64(x))
}

func (x View) Equal(y View) bool {
	return x == y
}

func (x View) KeyForMap() uint64 {
	return uint64(x)
}

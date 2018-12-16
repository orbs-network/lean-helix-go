// AUTO GENERATED FILE (by membufc proto compiler v0.0.21)
package primitives

import (
	"fmt"
	"bytes"
)

type MemberId []byte

func (x MemberId) String() string {
  return fmt.Sprintf("%x", []byte(x))
}

func (x MemberId) Equal(y MemberId) bool {
  return bytes.Equal(x, y)
}

func (x MemberId) KeyForMap() string {
  return string(x)
}

type Signature []byte

func (x Signature) String() string {
  return fmt.Sprintf("%x", []byte(x))
}

func (x Signature) Equal(y Signature) bool {
  return bytes.Equal(x, y)
}

func (x Signature) KeyForMap() string {
  return string(x)
}

type RandomSeedSignature []byte

func (x RandomSeedSignature) String() string {
  return fmt.Sprintf("%x", []byte(x))
}

func (x RandomSeedSignature) Equal(y RandomSeedSignature) bool {
  return bytes.Equal(x, y)
}

func (x RandomSeedSignature) KeyForMap() string {
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

type BlockHash []byte

func (x BlockHash) String() string {
  return fmt.Sprintf("%x", []byte(x))
}

func (x BlockHash) Equal(y BlockHash) bool {
  return bytes.Equal(x, y)
}

func (x BlockHash) KeyForMap() string {
  return string(x)
}



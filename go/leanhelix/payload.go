package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/go/block"
)

type Payload struct {
	PublicKey []byte
	Signature string
}
type PrepreparePayload struct {
	Payload
	Data  *PrepreparePayloadData
	Block *block.Block
}
type PrepreparePayloadData struct {
	MessageType MessageType
	BlockHash   []byte
	View        uint64
	Term        uint64
}

type PreparePayload struct {
	Payload
	Data *PreparePayloadData
}
type PreparePayloadData struct {
	MessageType MessageType
	BlockHash   []byte
	View        uint64
	Term        uint64
}

type CommitPayload struct {
	Payload
	Data *CommitPayloadData
}
type CommitPayloadData struct {
	MessageType MessageType
	BlockHash   []byte
	View        uint64
	Term        uint64
}

type ViewChangePayload struct {
	Payload
	Data ViewChangePayloadData
}
type ViewChangePayloadData struct {
	MessageType MessageType
	Term        uint64
	NewView     uint64
	PreparedProof
}

type NewViewPayload struct {
	Payload
	Data *NewViewPayloadData
}

type NewViewPayloadData struct {
	PrepreparePayload
	MessageType     MessageType
	ViewChangeProof []ViewChangePayload
	Term            uint64
	View            uint64
}

type PreparedProof struct {
	PrepreparePayload
	PreparePayloads []PreparePayload
}

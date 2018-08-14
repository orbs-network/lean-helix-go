package builders

import (
	"github.com/orbs-network/lean-helix-go/go/block"
	"github.com/orbs-network/lean-helix-go/go/networkcommunication"
	"github.com/orbs-network/lean-helix-go/go/test/blockutils"
	"github.com/orbs-network/lean-helix-go/go/test/keymanagermock"
)

func CreatePrepreparePayload(km *keymanagermock.KeyManagerMock, term uint64, view uint64, block *block.Block) *networkcommunication.PrepreparePayload {
	blockHash := blockutils.CalculateBlockHash(block)

	prepreparePayloadData := &networkcommunication.PrepreparePayloadData{
		BlockHash: blockHash,
		View:      view,
		Term:      term,
	}

	result := &networkcommunication.PrepreparePayload{
		Payload: networkcommunication.Payload{
			PublicKey: km.MyPublicKey(),
			Signature: km.Sign(prepreparePayloadData),
		},
		Data:  prepreparePayloadData,
		Block: block,
	}
	return result
}

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
			Signature: km.SignPrepreparePayloadData(prepreparePayloadData),
		},
		Data:  prepreparePayloadData,
		Block: block,
	}
	return result
}

func CreatePreparePayload(km *keymanagermock.KeyManagerMock, term uint64, view uint64, block *block.Block) *networkcommunication.PreparePayload {
	blockHash := blockutils.CalculateBlockHash(block)

	preparePayloadData := &networkcommunication.PreparePayloadData{
		BlockHash: blockHash,
		View:      view,
		Term:      term,
	}

	result := &networkcommunication.PreparePayload{
		Payload: networkcommunication.Payload{
			PublicKey: km.MyPublicKey(),
			Signature: km.SignPreparePayloadData(preparePayloadData),
		},
		Data: preparePayloadData,
	}
	return result

}

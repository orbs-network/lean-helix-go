package builders

import (
	"github.com/orbs-network/lean-helix-go/go/block"
	"github.com/orbs-network/lean-helix-go/go/leanhelix"
)

func CreatePrepreparePayload(km leanhelix.KeyManager, term uint64, view uint64, block *block.Block) *leanhelix.PrepreparePayload {
	blockHash := blockUtils.CalculateBlockHash(block)

	prepreparePayloadData := &leanhelix.PrepreparePayloadData{
		MessageType: leanhelix.MESSAGE_TYPE_PREPREPARE,
		BlockHash:   blockHash,
		View:        view,
		Term:        term,
	}

	result := &leanhelix.PrepreparePayload{
		Payload: leanhelix.Payload{
			PublicKey: km.MyPublicKey(),
			Signature: km.SignPrepreparePayloadData(prepreparePayloadData),
		},
		Data:  prepreparePayloadData,
		Block: block,
	}
	return result
}

func CreatePreparePayload(km leanhelix.KeyManager, term uint64, view uint64, block *block.Block) *leanhelix.PreparePayload {
	blockHash := blockUtils.CalculateBlockHash(block)

	preparePayloadData := &leanhelix.PreparePayloadData{
		MessageType: leanhelix.MESSAGE_TYPE_PREPARE,
		BlockHash:   blockHash,
		View:        view,
		Term:        term,
	}

	result := &leanhelix.PreparePayload{
		Payload: leanhelix.Payload{
			PublicKey: km.MyPublicKey(),
			Signature: km.SignPreparePayloadData(preparePayloadData),
		},
		Data: preparePayloadData,
	}
	return result

}

func CreateCommitPayload(km leanhelix.KeyManager, term uint64, view uint64, block *block.Block) *leanhelix.CommitPayload {
	blockHash := blockUtils.CalculateBlockHash(block)

	commitPayloadData := &leanhelix.CommitPayloadData{
		MessageType: leanhelix.MESSAGE_TYPE_COMMIT,
		BlockHash:   blockHash,
		View:        view,
		Term:        term,
	}

	result := &leanhelix.CommitPayload{
		Payload: leanhelix.Payload{
			PublicKey: km.MyPublicKey(),
			Signature: km.SignCommitPayloadData(commitPayloadData),
		},
		Data: commitPayloadData,
	}
	return result

}

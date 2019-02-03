package preparedmessages

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type PreparedMessages struct {
	PreprepareMessage *interfaces.PreprepareMessage
	PrepareMessages   []*interfaces.PrepareMessage
}

func ExtractPreparedMessages(blockHeight primitives.BlockHeight, latestPreparedView primitives.View, storage interfaces.Storage, q int) *PreparedMessages {

	ppm, ok := storage.GetPreprepareFromView(blockHeight, latestPreparedView)
	if !ok {
		return nil
	}

	prepareMessages, ok := storage.GetPrepareMessages(blockHeight, latestPreparedView, ppm.Content().SignedHeader().BlockHash())
	if !ok {
		return nil
	}

	if len(prepareMessages) < q-1 {
		return nil
	}

	return &PreparedMessages{
		PreprepareMessage: ppm,
		PrepareMessages:   prepareMessages,
	}
}

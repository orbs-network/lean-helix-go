package leanhelix

import "github.com/orbs-network/lean-helix-go/primitives"

type PreparedMessages struct {
	PreprepareMessage *PreprepareMessage
	PrepareMessages   []*PrepareMessage
}

func ExtractPreparedMessages(blockHeight primitives.BlockHeight, storage Storage, f int) *PreparedMessages {
	ppm, ok := storage.GetLatestPreprepare(blockHeight)
	if !ok {
		return nil
	}
	lastView := ppm.View()
	prepareMessages, ok := storage.GetPrepares(blockHeight, lastView, ppm.Content().SignedHeader().BlockHash())
	if !ok {
		return nil
	}
	if len(prepareMessages) >= 2*f {
		return &PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages:   prepareMessages,
		}
	}
	return nil
}

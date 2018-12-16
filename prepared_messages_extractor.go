package leanhelix

import "github.com/orbs-network/lean-helix-go/spec/types/go/primitives"

type PreparedMessages struct {
	PreprepareMessage *PreprepareMessage
	PrepareMessages   []*PrepareMessage
}

func ExtractPreparedMessages(blockHeight primitives.BlockHeight, storage Storage, q int) *PreparedMessages {
	ppm, ok := storage.GetLatestPreprepare(blockHeight)
	if !ok {
		return nil
	}
	lastView := ppm.View()
	prepareMessages, ok := storage.GetPrepareMessages(blockHeight, lastView, ppm.Content().SignedHeader().BlockHash())
	if !ok {
		return nil
	}
	if len(prepareMessages) >= q-1 {
		return &PreparedMessages{
			PreprepareMessage: ppm,
			PrepareMessages:   prepareMessages,
		}
	}
	return nil
}

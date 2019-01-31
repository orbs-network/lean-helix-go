package preparedmessages

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
)

type PreparedMessages struct {
	PreprepareMessage *interfaces.PreprepareMessage
	PrepareMessages   []*interfaces.PrepareMessage
}

func ExtractPreparedMessages(blockHeight primitives.BlockHeight, lastView primitives.View, storage interfaces.Storage, q int) *PreparedMessages {

	// TODO Change impl - loop from view-1 -> 0 and find a view which was PREPARED (P+PPs) and return that
	for v := lastView; v >= 0; v-- {
		if v == 0 { // SHITTY GO
			break
		}
		ppm, ok := storage.GetPreprepareFromView(blockHeight, v)
		if !ok {
			continue
		}

		prepareMessages, ok := storage.GetPrepareMessages(blockHeight, v, ppm.Content().SignedHeader().BlockHash())
		if !ok {
			continue
		}
		if len(prepareMessages) >= q-1 {
			return &PreparedMessages{
				PreprepareMessage: ppm,
				PrepareMessages:   prepareMessages,
			}
		}
	}
	return nil
}

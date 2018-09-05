package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/inmemoryblockchain"
	"github.com/orbs-network/lean-helix-go/types"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestMessageFactory(t *testing.T) {
	keyManager := lh.NewMockKeyManager("My PK")
	term := types.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := types.ViewCounter(math.Floor(rand.Float64() * 1000000))
	block := CreateBlock(inmemoryblockchain.GenesisBlock)
	blockHash := lh.CalculateBlockHash(block)
	messageFactory := lh.NewMessageFactory(lh.CalculateBlockHash, keyManager)

	t.Run("Construct Preprepare message", func(t *testing.T) {
		content := &lh.BlockMessageContent{
			MessageType: lh.MESSAGE_TYPE_PREPREPARE,
			Term:        term,
			View:        view,
			BlockHash:   blockHash,
		}
		expectedMessage := &lh.PrePrepareMessage{
			BlockRefMessage: &lh.BlockRefMessage{
				SignaturePair: &lh.SignaturePair{
					SignerPublicKey:  keyManager.MyPublicKey(),
					ContentSignature: keyManager.SignBlockMessageContent(content),
				},
				Content: content,
			},
			Block: block,
		}
		actualMessage := messageFactory.CreatePreprepareMessage(term, view, block)
		require.Equal(t, expectedMessage, actualMessage, "Preprepare message created")
	})

}

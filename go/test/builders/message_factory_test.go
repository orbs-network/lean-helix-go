package builders

import (
	lh "github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/test/inmemoryblockchain"
	"github.com/orbs-network/lean-helix-go/go/test/keymanagermock"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestMessageFactory(t *testing.T) {
	keyManager := keymanagermock.NewMockKeyManager("My PK")
	term := lh.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := lh.ViewCounter(math.Floor(rand.Float64() * 1000000))
	block := CreateBlock(inmemoryblockchain.GenesisBlock)
	blockHash := CalculateBlockHash(block)
	messageFactory := lh.NewMessageFactory(CalculateBlockHash, keyManager)

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

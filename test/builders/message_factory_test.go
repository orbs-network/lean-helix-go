package builders

import (
	lh "github.com/orbs-network/lean-helix-go"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestMessageFactory(t *testing.T) {
	keyManager := NewMockKeyManager(lh.PublicKey("My PK"))
	term := lh.BlockHeight(math.Floor(rand.Float64() * 1000000))
	view := lh.ViewCounter(math.Floor(rand.Float64() * 1000000))
	block := CreateBlock(GenesisBlock)
	blockHash := CalculateBlockHash(block)
	messageFactory := NewMessageFactory(CalculateBlockHash, keyManager)

	t.Run("Construct Preprepare message", func(t *testing.T) {
		content := &blockRef{
			messageType: lh.MESSAGE_TYPE_PREPREPARE,
			term:        term,
			view:        view,
			blockHash:   blockHash,
		}
		//expectedMessage := &lh.PrePrepareMessage{
		//	BlockRefMessage: &lh.BlockRefMessage{
		//		SignaturePair: &lh.SignaturePair{
		//			SignerPublicKey:  keyManager.MyID(),
		//			ContentSignature: keyManager.SignBlockMessageContent(content),
		//		},
		//		Content: content,
		//	},
		//	Block: block,
		//}

		ex := &preprepareMessage{
			blockRef: content,
			sender:   keyManager.SignBlockRef(content),
			block:    block,
		}
		actualMessage := messageFactory.CreatePreprepareMessage(term, view, block)
		require.Equal(t, ex, actualMessage, "Preprepare message created")
	})

}

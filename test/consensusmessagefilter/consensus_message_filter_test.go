package consensusmessagefilter

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func testLogger() leanhelix.Logger {
	return leanhelix.NewSilentLogger()
}

func GenerateMessage(blockHeight primitives.BlockHeight, view primitives.View, senderPublicKey string) leanhelix.ConsensusRawMessage {
	keyManager := builders.NewMockKeyManager(primitives.Ed25519PublicKey(senderPublicKey))
	block := builders.CreateBlock(builders.GenesisBlock)
	return builders.APrepareMessage(keyManager, blockHeight, view, block).ToConsensusRawMessage()
}

func TestGettingAMessage(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.Ed25519PublicKey("My PublicKey"), testLogger())
		rawMessage := GenerateMessage(10, 20, "Sender PublicKey")
		go filter.OnGossipMessage(ctx, rawMessage)

		actual, _ := filter.WaitForMessage(ctx, 10)
		expected := rawMessage.ToConsensusMessage()
		require.Equal(t, expected.Raw(), actual.Raw())
	})
}

func TestStoppingOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	filter := leanhelix.NewConsensusMessageFilter(primitives.Ed25519PublicKey("My PublicKey"), testLogger())
	rawMessage := GenerateMessage(10, 20, "Sender PublicKey")
	go filter.OnGossipMessage(ctx, rawMessage)

	actual, err := filter.WaitForMessage(ctx, 10)

	require.Nil(t, actual)
	require.Error(t, err)
}

func TestFilterMessagesFromThePast(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.Ed25519PublicKey("My PublicKey"), testLogger())
		rawMessageFromThePast := GenerateMessage(9, 20, "Sender PublicKey")
		rawMessageFromThePresent := GenerateMessage(10, 20, "Sender PublicKey")
		go func() {
			filter.OnGossipMessage(ctx, rawMessageFromThePast)
			filter.OnGossipMessage(ctx, rawMessageFromThePresent)
		}()

		actual, _ := filter.WaitForMessage(ctx, 10)
		expected := rawMessageFromThePresent.ToConsensusMessage()
		require.Equal(t, expected.Raw(), actual.Raw())
	})
}

func TestCacheMessagesFromTheFuture(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.Ed25519PublicKey("My PublicKey"), testLogger())
		rawMessageFromTheFuture := GenerateMessage(11, 20, "Sender PublicKey")
		rawMessageFromThePresent := GenerateMessage(10, 20, "Sender PublicKey")
		go func() {
			filter.OnGossipMessage(ctx, rawMessageFromTheFuture)
			filter.OnGossipMessage(ctx, rawMessageFromThePresent)
		}()

		actualOn10, _ := filter.WaitForMessage(ctx, 10)
		expectedOn10 := rawMessageFromThePresent.ToConsensusMessage()
		require.Equal(t, expectedOn10.Raw(), actualOn10.Raw())

		actualOn11, _ := filter.WaitForMessage(ctx, 11)
		expectedOn11 := rawMessageFromTheFuture.ToConsensusMessage()
		require.Equal(t, expectedOn11.Raw(), actualOn11.Raw())
	})
}

func TestFilterMessagesWithMyPublicKey(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.Ed25519PublicKey("My PublicKey"), testLogger())
		badMessage := GenerateMessage(10, 20, "My PublicKey")
		goodMessage := GenerateMessage(10, 20, "Sender PublicKey")
		go func() {
			filter.OnGossipMessage(ctx, badMessage)
			filter.OnGossipMessage(ctx, goodMessage)
		}()

		actual, _ := filter.WaitForMessage(ctx, 10)
		expected := goodMessage.ToConsensusMessage()
		require.Equal(t, expected.Raw(), actual.Raw())
	})
}

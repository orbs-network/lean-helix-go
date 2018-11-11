package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

func GenerateMessage(blockHeight primitives.BlockHeight, view primitives.View, senderPublicKey string) leanhelix.ConsensusRawMessage {
	keyManager := builders.NewMockKeyManager(primitives.Ed25519PublicKey(senderPublicKey))
	block := builders.CreateBlock(builders.GenesisBlock)
	return builders.APrepareMessage(keyManager, blockHeight, view, block).ToConsensusRawMessage()
}

func TestGettingAMessage(t *testing.T) {
	WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.Ed25519PublicKey("My PublicKey"))
		rawMessage := GenerateMessage(10, 20, "Sender PublicKey")
		go filter.OnGossipMessage(rawMessage)

		actual, _ := filter.WaitForMessage(ctx, 10)
		expected := rawMessage.ToConsensusMessage()
		require.Equal(t, expected.Raw(), actual.Raw())
	})
}

func TestStoppingOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	filter := leanhelix.NewConsensusMessageFilter(primitives.Ed25519PublicKey("My PublicKey"))
	rawMessage := GenerateMessage(10, 20, "Sender PublicKey")
	go filter.OnGossipMessage(rawMessage)

	actual, err := filter.WaitForMessage(ctx, 10)

	require.Nil(t, actual)
	require.Error(t, err)
}

func TestFilterMessagesFromThePast(t *testing.T) {
	WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.Ed25519PublicKey("My PublicKey"))
		rawMessageFromThePast := GenerateMessage(9, 20, "Sender PublicKey")
		rawMessageFromThePresent := GenerateMessage(10, 20, "Sender PublicKey")
		go func() {
			filter.OnGossipMessage(rawMessageFromThePast)
			filter.OnGossipMessage(rawMessageFromThePresent)
		}()

		actual, _ := filter.WaitForMessage(ctx, 10)
		expected := rawMessageFromThePresent.ToConsensusMessage()
		require.Equal(t, expected.Raw(), actual.Raw())
	})
}

func TestCacheMessagesFromTheFuture(t *testing.T) {
	WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.Ed25519PublicKey("My PublicKey"))
		rawMessageFromTheFuture := GenerateMessage(11, 20, "Sender PublicKey")
		rawMessageFromThePresent := GenerateMessage(10, 20, "Sender PublicKey")
		go func() {
			filter.OnGossipMessage(rawMessageFromTheFuture)
			filter.OnGossipMessage(rawMessageFromThePresent)
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
	WithContext(func(ctx context.Context) {
		filter := leanhelix.NewConsensusMessageFilter(primitives.Ed25519PublicKey("My PublicKey"))
		badMessage := GenerateMessage(10, 20, "My PublicKey")
		goodMessage := GenerateMessage(10, 20, "Sender PublicKey")
		go func() {
			filter.OnGossipMessage(badMessage)
			filter.OnGossipMessage(goodMessage)
		}()

		actual, _ := filter.WaitForMessage(ctx, 10)
		expected := goodMessage.ToConsensusMessage()
		require.Equal(t, expected.Raw(), actual.Raw())
	})
}

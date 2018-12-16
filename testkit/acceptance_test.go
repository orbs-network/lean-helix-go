package testkit

// Shai comments:
// One: mocks
// Two: Actual impl

// Use MemoryTransport
//

// TODO uncomment and use spi struct
//func TestHappyFlow(t *testing.T, spi *protocol.LeanHelixSPI) {
//
//	test.WithContext(func(ctx context.Context) {
//		block1 := builders.CreateBlock(builders.GenesisBlock)
//		block2 := builders.CreateBlock(block1)
//
//		//testNetwork := builders.ATestNetwork(ctx, 4, block1, block2)
//		testNetwork := builders.CreateTestNetworkForConsumerTests(ctx, 4, spi, block1, block2)
//		testNetwork.StartConsensus(ctx)
//
//		require.True(t, testNetwork.WaitForAllNodesToCommitBlock(block1))
//	})
//}

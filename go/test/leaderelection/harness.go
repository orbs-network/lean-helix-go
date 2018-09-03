package leaderelection

//type TriggerElectionFunc func([]ElectionTrigger)
//
//type harness struct {
//	TestNetwork TestNetwork
//	blockUtils leanhelix.BlockUtils
//	blocksPool []leanhelix.Block
//	triggerElection
//
//}
//
//
//func NewHarness(nodeCount int) {
//
//	blockUtils := builders.NewMockBlockUtils()
//	triggerElection := func () {
//		for _, t := range electionTriggers {
//			t.Trigger()
//		}
//	}
//
//	testNetwork := builders.CreateTestNetwork(nodeCount)
//	.electingLeaderUsing(electionTriggerFactory)
//	.gettingBlocksVia(blockUtils)
//	.thatLogsToCustomeLogger(SocketsLogger)
//	.with(countOfNodes).nodes
//	.build()
//
//
//
//}

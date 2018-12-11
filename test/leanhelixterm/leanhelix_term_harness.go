package leanhelixterm

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/primitives"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"github.com/stretchr/testify/require"
	"testing"
)

type harness struct {
	t                 *testing.T
	myPublicKey       primitives.Ed25519PublicKey
	keyManager        *builders.MockKeyManager
	myNode            *builders.Node
	net               *builders.TestNetwork
	term              *leanhelix.LeanHelixTerm
	storage           leanhelix.Storage
	electionTrigger   *builders.ElectionTriggerMock
	failVerifications bool
}

func NewHarness(ctx context.Context, t *testing.T, blocksPool ...leanhelix.Block) *harness {
	net := builders.NewTestNetworkBuilder().WithNodeCount(4).WithBlocks(blocksPool).Build()
	myNode := net.Nodes[0]
	keyManager := myNode.KeyManager
	termConfig := myNode.BuildConfig()
	term := leanhelix.NewLeanHelixTerm(ctx, termConfig, nil, myNode.GetLatestBlock())
	term.StartTerm(ctx)

	return &harness{
		t:                 t,
		myPublicKey:       keyManager.MyPublicKey(),
		myNode:            myNode,
		net:               net,
		keyManager:        myNode.KeyManager,
		term:              term,
		storage:           termConfig.Storage,
		electionTrigger:   myNode.ElectionTrigger,
		failVerifications: false,
	}
}

func (h *harness) failValidations() {
	h.myNode.BlockUtils.ValidationResult = false
}

func (h *harness) checkView(expectedView primitives.View) {
	view := h.term.GetView()
	require.Equal(h.t, expectedView, view, fmt.Sprintf("Term should have view=%d, but got %d", expectedView, view))
}

func (h *harness) triggerElection(ctx context.Context) {
	h.electionTrigger.ManualTriggerSync(ctx)
}

func (h *harness) getMyNodePk() primitives.Ed25519PublicKey {
	return h.getMemberPk(0)
}

func (h *harness) getMemberPk(nodeIdx int) primitives.Ed25519PublicKey {
	return h.net.Nodes[nodeIdx].KeyManager.MyPublicKey()
}

func (h *harness) getMyKeyManager() leanhelix.KeyManager {
	return h.getMemberKeyManager(0)
}

func (h *harness) getMemberKeyManager(nodeIdx int) leanhelix.KeyManager {
	return h.net.Nodes[nodeIdx].KeyManager
}

func (h *harness) getMembersKeyManagers(expectNodeIdx int) []leanhelix.KeyManager {
	var keyManagers []leanhelix.KeyManager
	for i, node := range h.net.Nodes {
		if i != expectNodeIdx {
			keyManagers = append(keyManagers, node.KeyManager)
		}
	}

	return keyManagers
}

func (h *harness) electionTillView(ctx context.Context, view primitives.View) {
	for {
		if h.term.GetView() == view {
			break
		}
		h.triggerElection(ctx)
	}
}

func (h *harness) setNode1AsTheLeader(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	h.receiveNewView(ctx, 1, blockHeight, view, block)
}

func (h *harness) setMeAsTheLeader(ctx context.Context, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	h.receiveNewView(ctx, 0, blockHeight, view, block)
}

func (h *harness) receiveViewChange(ctx context.Context, fromNodeIdx int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	sender := h.net.Nodes[fromNodeIdx]
	vc := builders.AViewChangeMessage(sender.KeyManager, blockHeight, view, nil)
	h.term.HandleLeanHelixViewChange(ctx, vc)
}

func (h *harness) receiveViewChangeMessage(ctx context.Context, msg *leanhelix.ViewChangeMessage) {
	h.term.HandleLeanHelixViewChange(ctx, msg)
}

func (h *harness) receivePreprepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	leader := h.net.Nodes[fromNode]
	ppm := builders.APreprepareMessage(leader.KeyManager, blockHeight, view, block)
	h.term.HandleLeanHelixPrePrepare(ctx, ppm)
}

func (h *harness) receivePreprepareMessage(ctx context.Context, ppm *leanhelix.PreprepareMessage) {
	h.term.HandleLeanHelixPrePrepare(ctx, ppm)
}

func (h *harness) receivePrepare(ctx context.Context, fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	sender := h.net.Nodes[fromNode]
	pm := builders.APrepareMessage(sender.KeyManager, blockHeight, view, block)
	h.term.HandleLeanHelixPrepare(ctx, pm)
}

func (h *harness) createPreprepareMessage(fromNode int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block, blockHash primitives.Uint256) *leanhelix.PreprepareMessage {
	leader := h.net.Nodes[fromNode]
	messageFactory := leanhelix.NewMessageFactory(leader.KeyManager)
	return messageFactory.CreatePreprepareMessage(blockHeight, view, block, blockHash)
}

func (h *harness) HandleLeanHelixNewView(ctx context.Context, nvm *leanhelix.NewViewMessage) {
	h.term.HandleLeanHelixNewView(ctx, nvm)
}

func (h *harness) receiveNewView(ctx context.Context, fromNodeIdx int, blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) {
	leaderKeyManager := h.getMemberKeyManager(fromNodeIdx)
	membersKeyManagers := h.getMembersKeyManagers(fromNodeIdx)
	votes := builders.ASimpleViewChangeVotes(membersKeyManagers, blockHeight, view)
	nvm := builders.
		NewNewViewBuilder().
		LeadBy(leaderKeyManager).
		WithViewChangeVotes(votes).
		OnBlock(block).
		OnBlockHeight(blockHeight).
		OnView(view).
		Build()
	h.term.HandleLeanHelixNewView(ctx, nvm)
}

func (h *harness) getLastSentViewChangeMessage() *leanhelix.ViewChangeMessage {
	messages := h.myNode.Gossip.GetSentMessages(leanhelix.LEAN_HELIX_VIEW_CHANGE)
	lastMessage := messages[len(messages)-1].ToConsensusMessage()
	return lastMessage.(*leanhelix.ViewChangeMessage)
}

func (h *harness) countViewChange(blockHeight primitives.BlockHeight, view primitives.View) int {
	messages, _ := h.storage.GetViewChangeMessages(blockHeight, view)
	return len(messages)
}

func (h *harness) countCommits(blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) int {
	messages, _ := h.storage.GetCommitMessages(blockHeight, view, builders.CalculateBlockHash(block))
	return len(messages)
}

func (h *harness) hasPreprepare(blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) bool {
	message, ok := h.storage.GetPreprepareMessage(blockHeight, view)

	if message == nil || !ok {
		return false
	}

	return builders.BlocksAreEqual(message.Block(), block)
}

func (h *harness) countPrepare(blockHeight primitives.BlockHeight, view primitives.View, block leanhelix.Block) int {
	messages, _ := h.storage.GetPrepareMessages(blockHeight, view, builders.CalculateBlockHash(block))
	return len(messages)
}

func (h *harness) failFutureVerifications() {
	h.keyManager.FailFutureVerifications = true
}

func (h *harness) disposeTerm() {
	h.term.Dispose()
}

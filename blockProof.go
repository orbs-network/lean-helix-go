package leanhelix

import (
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
)

// assume commit messages are valid and still hold
func GenerateLeanHelixBlockProof(commitMessages []*CommitMessage) *protocol.BlockProof {
	blockHeight := commitMessages[0].BlockHeight()
	blockRefBuilder := &protocol.BlockRefBuilder{
		MessageType: protocol.LEAN_HELIX_COMMIT,
		BlockHeight: blockHeight,
		View:        commitMessages[0].View(),
		BlockHash:   commitMessages[0].Content().SignedHeader().BlockHash(),
	}

	cSendersBuilders := make([]*protocol.SenderSignatureBuilder, 0)
	for _, cm := range commitMessages {
		memberId := cm.Content().Sender().MemberId()
		cSendersBuilders = append(cSendersBuilders, &protocol.SenderSignatureBuilder{
			MemberId:  memberId,
			Signature: cm.Content().Sender().Signature(),
		})
	}

	return (&protocol.BlockProofBuilder{
		BlockRef: blockRefBuilder,
		Nodes:    cSendersBuilders,
	}).Build()
}

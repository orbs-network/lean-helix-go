package impl

import (
	"github.com/orbs-network/lean-helix-go"
)

// Reference implementation of KeyManagerImpl

type KeyManagerReferenceImpl struct {
}

func (mgr *KeyManagerReferenceImpl) SignBlockRef(blockRef leanhelix.BlockRef) leanhelix.SenderSignature {
	panic("implement me")
}

func (mgr *KeyManagerReferenceImpl) SignViewChange(vcHeader leanhelix.ViewChangeHeader) leanhelix.SenderSignature {
	panic("implement me")
}

func (mgr *KeyManagerReferenceImpl) SignNewView(nvHeader leanhelix.NewViewHeader) leanhelix.SenderSignature {

	bytes := serialize(nvHeader) // Reader, Builder in membuffers
	signature := sign(bytes, privateKey)

	return NewSenderSignature(sender, signature)

	panic("implement me")
}

func (mgr *KeyManagerReferenceImpl) VerifyBlockRef(blockRef leanhelix.BlockRef, sender leanhelix.SenderSignature) bool {
	// From TS: return this.verify(blockRef, sender);
	// From spec:

	sender.SenderPublicKey()
}

func (mgr *KeyManagerReferenceImpl) VerifyViewChange(vcHeader leanhelix.ViewChangeHeader, sender leanhelix.SenderSignature) bool {
	panic("implement me")
}

func (mgr *KeyManagerReferenceImpl) VerifyNewView(nvHeader leanhelix.NewViewHeader, sender leanhelix.SenderSignature) bool {

	/*
							Discard if the sender is not a valid participant node
							Discard if the sender is the leader for the view based on GetCurrentLeader
							Discard if message.view < my_state.view
							Discard if a PRE_PREPARE message was already logged for the same view and sender (NVPP?)

		VerifySignature(nvHeader)





					Verify 2f+1 VIEW_CHANGE messages, from different senders
			--- nvHeader.ViewChangeConfirmations()[0].Sender().SenderPublicKey()

				For each VIEW_CHANGE message verify:
				Type = VIEW_CHANGE
				Sender is a valid participant node
				Block_height = NEW_VIEW message.Block_height
				View = NEW_VIEW message.View
				Prepared_proof is valid using ValidatePreparedProof(View_change_view)
				Valid signature
				Discard if one of the checks fails.

				Check the New View PRE_PREPARE message fields
			Check Type = PRE_PREPARE
			Check Sender = NEW_VIEW.Sender
			Check View = NEW_VIEW.View
			Check Block_height = NEW_VIEW.Block_height
			Compre the Block_hash to the hash of teh NewView message.block calculated by calling ConsensusAlgo.CalcBlockHash.
			Check the NVPP signature



	*/

	panic("implement me")
}

func (mgr *KeyManagerReferenceImpl) MyPublicKey() leanhelix.PublicKey {
	panic("implement me")
}

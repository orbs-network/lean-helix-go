package builders

//type MockPreparedProof struct {
//	ppBlockRef lh.BlockRef
//	pBlockRef  lh.BlockRef
//	ppSigner   lh.SenderSignature
//	pSigners   []lh.SenderSignature
//}
//
//func (p *MockPreparedProof) PPBlockRef() lh.BlockRef {
//	return p.ppBlockRef
//}
//
//func (p *MockPreparedProof) PBlockRef() lh.BlockRef {
//	return p.pBlockRef
//}
//
//func (p *MockPreparedProof) PPSender() lh.SenderSignature {
//	return p.ppSigner
//}
//
//func (p *MockPreparedProof) PSenders() []lh.SenderSignature {
//	return p.pSigners
//}
//
//func CreateMockPreparedProof(preprepareMessage lh.PreprepareMessage, prepareMessages []lh.PrepareMessage) lh.PreparedProof {
//
//	var (
//		pBlockRef lh.BlockRef
//		pSigners  []lh.SenderSignature
//	)
//
//	ppBlockRef := preprepareMessage.SignedHeader()
//	ppSigner := preprepareMessage.Sender()
//
//	if len(prepareMessages) > 0 {
//		pBlockRef = prepareMessages[0].SignedHeader()
//		pSigners = make([]lh.SenderSignature, 0, len(prepareMessages))
//		for _, pm := range prepareMessages {
//			pSigners = append(pSigners, pm.Sender())
//		}
//	} else {
//		pBlockRef = nil
//		pSigners = nil
//	}
//
//	return CreateMockPreparedProof_(ppBlockRef, pBlockRef, ppSigner, pSigners)
//}
//
//func CreateMockPreparedProof_(
//	ppBlockRef lh.BlockRef,
//	pBlockRef lh.BlockRef,
//	ppSigner lh.SenderSignature,
//	pSigners []lh.SenderSignature) lh.PreparedProof {
//	return &MockPreparedProof{
//		ppBlockRef,
//		pBlockRef,
//		ppSigner,
//		pSigners,
//	}
//}

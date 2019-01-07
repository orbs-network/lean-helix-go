package randomseed

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"strconv"
)

func RandomSeedToBytes(randomSeed uint64) []byte {
	return []byte(strconv.FormatUint(randomSeed, 10))
}

func CalculateRandomSeed(signature []byte) uint64 {
	hash := sha256.Sum256(signature)
	array := []byte{hash[0], hash[3], hash[7], hash[11], hash[15], hash[19], hash[23], hash[27]}
	return binary.LittleEndian.Uint64(array)
}

func ValidateRandomSeed(keyManager interfaces.KeyManager, blockHeight primitives.BlockHeight, blockProof *protocol.BlockProof, prevBlockProof *protocol.BlockProof) bool {
	randomSeedSignature := blockProof.RandomSeedSignature()
	prevRandomSeedSignature := prevBlockProof.RandomSeedSignature()

	randomSeed := CalculateRandomSeed(prevRandomSeedSignature) // Calculate the random seed based on prev block proof

	// validate random seed signature against master publicKey
	masterRandomSeed := (&protocol.SenderSignatureBuilder{
		Signature: primitives.Signature(randomSeedSignature),
		MemberId:  nil, // master
	}).Build()

	if !keyManager.VerifyRandomSeed(blockHeight, RandomSeedToBytes(randomSeed), masterRandomSeed) {
		return false
	}

	return true
}

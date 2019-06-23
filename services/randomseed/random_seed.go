// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package randomseed

import (
	"crypto/sha256"
	"encoding/binary"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/pkg/errors"
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

func ValidateRandomSeed(keyManager interfaces.KeyManager, blockHeight primitives.BlockHeight, blockProof *protocol.BlockProof, prevBlockProof *protocol.BlockProof) error {
	randomSeedSignature := blockProof.RandomSeedSignature()
	prevRandomSeedSignature := prevBlockProof.RandomSeedSignature()

	randomSeed := CalculateRandomSeed(prevRandomSeedSignature) // Calculate the random seed based on prev block proof

	// validate random seed signature against master publicKey
	masterRandomSeed := (&protocol.SenderSignatureBuilder{
		Signature: primitives.Signature(randomSeedSignature),
		MemberId:  nil, // master
	}).Build()

	if err := keyManager.VerifyRandomSeed(blockHeight, RandomSeedToBytes(randomSeed), masterRandomSeed); err != nil {
		return errors.Wrap(err, "VerifyRandomSeed() failed")
	}

	return nil
}

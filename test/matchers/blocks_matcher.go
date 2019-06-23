// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package matchers

import (
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	"github.com/orbs-network/lean-helix-go/test/mocks"
)

func BlocksAreEqual(block1 interfaces.Block, block2 interfaces.Block) bool {
	return mocks.CalculateBlockHash(block1).Equal(mocks.CalculateBlockHash(block2))
}

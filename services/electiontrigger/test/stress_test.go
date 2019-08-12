// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"context"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/test"
	"testing"
	"time"
)

func TestStress_FrequentRegisters(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		et := buildElectionTrigger(ctx, 1*time.Microsecond)

		for h := primitives.BlockHeight(1); h < primitives.BlockHeight(1000); h++ {
			et.RegisterOnElection(h, 0, nil)
			time.Sleep(1 * time.Microsecond)
		}
	})
}

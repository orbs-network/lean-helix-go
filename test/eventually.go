// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"time"
)

const eventuallyIterations = 50
const consistentlyIterations = 25

func Eventually(timeout time.Duration, f func() bool) bool {
	for i := 0; i < eventuallyIterations; i++ {
		if f() {
			return true
		}
		time.Sleep(timeout / eventuallyIterations)
	}
	return false
}

func Consistently(timeout time.Duration, f func() bool) bool {
	for i := 0; i < consistentlyIterations; i++ {
		if !f() {
			return false
		}
		time.Sleep(timeout / consistentlyIterations)
	}
	return true
}

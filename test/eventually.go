// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
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


func NameHashPrefix(tb testing.TB, idLen int) string {
	if tb == nil {
		return strings.Repeat(" ", idLen)
	}
	testInstance := fmt.Sprintf("%p", tb) // test instance identifier
	md5 := md5.Sum([]byte(testInstance)) // avoid collisions
	return hex.EncodeToString(md5[:])[:idLen]
}
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

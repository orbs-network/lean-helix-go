package test

import (
	"fmt"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/lean-helix-go"
	"github.com/orbs-network/lean-helix-go/test/builders"
	"testing"
)

// Reference implementation of KeyManagerImpl usage

func (k *MockKeyManager) Sign(content []byte) []byte {
	return []byte{'e', 'n', 'i', 'g', 'm', 'a'} // Put something here
}

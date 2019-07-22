package mocks

import "github.com/orbs-network/lean-helix-go/services/interfaces"

func NewMockConfig() *interfaces.Config {
	return &interfaces.Config{
		Logger: nil,
	}
}

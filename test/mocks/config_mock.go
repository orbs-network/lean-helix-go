package mocks

import "github.com/orbs-network/lean-helix-go/services/interfaces"

func NewMockConfig() *interfaces.Config {
	return &interfaces.Config{
		Logger:     nil,
		Membership: NewMockMembership([]byte{30, 30, 30}, nil, false),
	}
}

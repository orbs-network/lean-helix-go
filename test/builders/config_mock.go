package builders

import "github.com/orbs-network/lean-helix-go"

type mockConfig struct {
	electionTrigger leanhelix.ElectionTrigger
}

func (c *mockConfig) ElectionTrigger() leanhelix.ElectionTrigger {
	return c.electionTrigger
}

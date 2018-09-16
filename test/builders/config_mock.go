package builders

import lh "github.com/orbs-network/lean-helix-go"

type mockConfig struct {
	electionTrigger lh.ElectionTrigger
}

func (c *mockConfig) ElectionTrigger() lh.ElectionTrigger {
	return c.electionTrigger
}

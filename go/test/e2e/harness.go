package e2e

import (
	"github.com/orbs-network/lean-helix-go/go/leanhelix"
	"github.com/orbs-network/lean-helix-go/go/test/builders"
	"github.com/orbs-network/lean-helix-go/go/test/keymanagermock"
)

type harness struct {
	config     *leanhelix.Config
	network    leanhelix.NetworkCommunication
	blockUtils leanhelix.BlockUtils
	keyManager leanhelix.KeyManager
	service    leanhelix.Service
}

func NewHarness(nodeCount int) *harness {

	config := &leanhelix.Config{}
	network := builders.NewMockNetworkCommunication(nodeCount)
	blockUtils := builders.NewMockBlockUtils()
	keyManager := keymanagermock.NewMockKeyManager(nil, nil)

	s := leanhelix.NewLeanHelix(config, network, blockUtils, keyManager)

	return &harness{
		// Place here anything else that serves the tests but is not part of "service"
		service: s,
	}
}

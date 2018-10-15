package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/instrumentation/log"
)

type Config struct {
	Ctx                  context.Context
	NetworkCommunication NetworkCommunication
	BlockUtils           BlockUtils
	KeyManager           KeyManager
	Logger               log.BasicLogger
	ElectionTrigger      ElectionTrigger
	Storage              Storage
}

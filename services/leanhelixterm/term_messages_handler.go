// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package leanhelixterm

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
)

type TermMessagesHandler interface {
	HandlePrePrepare(ctx context.Context, ppm *interfaces.PreprepareMessage)
	HandlePrepare(ctx context.Context, pm *interfaces.PrepareMessage)
	HandleViewChange(ctx context.Context, vcm *interfaces.ViewChangeMessage)
	HandleCommit(ctx context.Context, cm *interfaces.CommitMessage)
	HandleNewView(ctx context.Context, nvm *interfaces.NewViewMessage)
}

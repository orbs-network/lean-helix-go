// Copyright 2019 the lean-helix-go authors
// This file is part of the lean-helix-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package mocks

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
)

type TermMessagesHandlerMock struct {
	HistoryPP []*interfaces.PreprepareMessage
	HistoryP  []*interfaces.PrepareMessage
	HistoryC  []*interfaces.CommitMessage
	HistoryNV []*interfaces.NewViewMessage
	HistoryVC []*interfaces.ViewChangeMessage
}

func NewTermMessagesHandlerMock() *TermMessagesHandlerMock {
	return &TermMessagesHandlerMock{}
}

func (tmh *TermMessagesHandlerMock) HandlePrePrepare(ctx context.Context, ppm *interfaces.PreprepareMessage) {
	tmh.HistoryPP = append(tmh.HistoryPP, ppm)
}

func (tmh *TermMessagesHandlerMock) HandlePrepare(ctx context.Context, pm *interfaces.PrepareMessage) {
	tmh.HistoryP = append(tmh.HistoryP, pm)
}

func (tmh *TermMessagesHandlerMock) HandleCommit(ctx context.Context, cm *interfaces.CommitMessage) {
	tmh.HistoryC = append(tmh.HistoryC, cm)
}

func (tmh *TermMessagesHandlerMock) HandleNewView(ctx context.Context, nvm *interfaces.NewViewMessage) {
	tmh.HistoryNV = append(tmh.HistoryNV, nvm)
}

func (tmh *TermMessagesHandlerMock) HandleViewChange(ctx context.Context, vcm *interfaces.ViewChangeMessage) {
	tmh.HistoryVC = append(tmh.HistoryVC, vcm)
}

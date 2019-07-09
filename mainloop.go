package leanhelix

import (
	"context"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"math"
)

type MainLoop struct {
	messagesChannel    chan *interfaces.ConsensusRawMessage
	updateStateChannel chan *blockWithProof
	currentHeight      primitives.BlockHeight
	config             *interfaces.Config
	logger             L.LHLogger
	onCommitCallback   interfaces.OnCommitCallback
	worker             *WorkerLoop
}

func NewLeanHelix(config *interfaces.Config, onCommitCallback interfaces.OnCommitCallback) *MainLoop {
	return &MainLoop{
		config:             config,
		onCommitCallback:   onCommitCallback,
		messagesChannel:    make(chan *interfaces.ConsensusRawMessage),
		updateStateChannel: make(chan *blockWithProof),
		currentHeight:      0,
		logger:             LoggerToLHLogger(config.Logger),
	}
}

func (m *MainLoop) Run(ctx context.Context) {

	workerCtx, cancelWorkerContext := context.WithCancel(ctx)
	m.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, m.config.Membership.MyMemberId()), "LHFLOW MAINLOOP START")
	m.RunWorkerLoop(ctx)
	m.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, m.config.Membership.MyMemberId()), "LHMSG START LISTENING NOW")
	for {
		select {

		case <-ctx.Done(): // system shutdown
			m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "LHFLOW MAINLOOP DONE, Terminating Run().")
			m.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, m.config.Membership.MyMemberId()), "LHMSG STOPPED LISTENING")
			return

		case message := <-m.messagesChannel:
			m.worker.MessagesChannel <- &MessageWithContext{ctx: workerCtx, msg: message}

		case trigger := <-m.config.ElectionTrigger.ElectionChannel():
			cancelWorkerContext()
			workerCtx, cancelWorkerContext = context.WithCancel(ctx)

			m.worker.ElectionChannel <- trigger

		case receivedBlockWithProof := <-m.updateStateChannel: // NodeSync
			cancelWorkerContext()
			workerCtx, cancelWorkerContext = context.WithCancel(ctx)

			m.worker.UpdateStateChannel <- receivedBlockWithProof

		}
	}
}

func (m *MainLoop) RunWorkerLoop(ctx context.Context) {

	lhLog := LoggerToLHLogger(m.config.Logger)
	m.worker = NewWorkerLoop(m.config, lhLog, m.onCommitCallback)

	go m.worker.Run(ctx)

}

func LoggerToLHLogger(logger interfaces.Logger) L.LHLogger {
	var lhLog L.LHLogger
	if logger == nil {
		lhLog = L.NewLhLogger(L.NewSilentLogger())
	} else {
		lhLog = L.NewLhLogger(logger)
	}

	return lhLog
}

func (m *MainLoop) GetCurrentHeight() primitives.BlockHeight {
	return m.currentHeight
}

func (m *MainLoop) UpdateState(ctx context.Context, prevBlock interfaces.Block, prevBlockProofBytes []byte) {
	if m.worker == nil {
		panic("UpdateState() worker is nil")
	}
	m.worker.UpdateState(ctx, prevBlock, prevBlockProofBytes)
}

func (m *MainLoop) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProofBytes []byte, maybePrevBlockProofBytes []byte) error {
	if m.worker == nil {
		panic("ValidateBlockConsensus() worker is nil")
	}
	return m.worker.ValidateBlockConsensus(ctx, block, blockProofBytes, maybePrevBlockProofBytes)
}

func (m *MainLoop) HandleConsensusMessage(ctx context.Context, message *interfaces.ConsensusRawMessage) {
	if m.worker == nil {
		panic("HandleConsensusMessage() worker is nil")
	}
	m.worker.HandleConsensusMessage(ctx, message)
}

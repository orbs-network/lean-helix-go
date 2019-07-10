package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"math"
)

type MainLoop struct {
	messagesChannel    chan *interfaces.ConsensusRawMessage
	updateStateChannel chan *blockWithProof
	electionChannel    chan func(ctx context.Context)

	currentHeight    primitives.BlockHeight
	config           *interfaces.Config
	logger           L.LHLogger
	onCommitCallback interfaces.OnCommitCallback

	worker *WorkerLoop
}

func NewLeanHelix(config *interfaces.Config, onCommitCallback interfaces.OnCommitCallback) *MainLoop {
	electionChannel := config.ElectionTrigger.ElectionChannel()

	return &MainLoop{
		config:             config,
		onCommitCallback:   onCommitCallback,
		messagesChannel:    make(chan *interfaces.ConsensusRawMessage, config.MsgChanBufLen),
		updateStateChannel: make(chan *blockWithProof, config.UpdateStateChanBufLen),
		electionChannel:    electionChannel,
		currentHeight:      0,
		logger:             LoggerToLHLogger(config.Logger),
	}
}

// ORBS: LeanHelix.Run(ctx, goroutineLauncher func(f func()) { GoForever(f) }))
// LH: goroutineLauncher(func (){m.RunWorkerLoop(ctx)})

func (m *MainLoop) Run(ctx context.Context) {
	m.RunWorkerLoop(ctx)
	m.RunMainLoop(ctx)
}

func (m *MainLoop) RunMainLoop(ctx context.Context) {

	go m.run(ctx)

}

func (m *MainLoop) run(ctx context.Context) {
	m.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, m.config.Membership.MyMemberId()), "LHFLOW MAINLOOP START")
	m.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, m.config.Membership.MyMemberId()), "LHMSG START LISTENING NOW")
	workerCtx, cancelWorkerContext := context.WithCancel(ctx)
	for {
		select {

		case <-ctx.Done(): // system shutdown

			m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "LHFLOW MAINLOOP DONE, Terminating Run().")
			m.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, m.config.Membership.MyMemberId()), "LHMSG STOPPED LISTENING")
			return

		case message := <-m.messagesChannel:
			fmt.Printf("%v Read from messages channel\n", m.config.Membership.MyMemberId())
			m.worker.MessagesChannel <- &MessageWithContext{ctx: workerCtx, msg: message}

		case trigger := <-m.electionChannel:
			fmt.Printf("%v Read from election channel\n", m.config.Membership.MyMemberId())
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

func (m *MainLoop) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProofBytes []byte, maybePrevBlockProofBytes []byte) error {
	if m.worker == nil {
		panic("ValidateBlockConsensus() worker is nil")
	}
	return m.worker.ValidateBlockConsensus(ctx, block, blockProofBytes, maybePrevBlockProofBytes)
}

// Called from outside to indicate Node Sync
func (m *MainLoop) UpdateState(ctx context.Context, prevBlock interfaces.Block, prevBlockProofBytes []byte) {

	select {
	case <-ctx.Done():
		m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "UpdateState() ID=%s CONTEXT TERMINATED", termincommittee.Str(m.config.Membership.MyMemberId()))
		return
	case m.updateStateChannel <- &blockWithProof{
		block:               prevBlock,
		prevBlockProofBytes: prevBlockProofBytes,
	}:
	}
}

// called by tests
func (m *MainLoop) TriggerElection(ctx context.Context, f func(ctx context.Context)) {
	select {
	case <-ctx.Done():
		m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "TriggerElection() ID=%s CONTEXT TERMINATED", termincommittee.Str(m.config.Membership.MyMemberId()))
		return
	case m.electionChannel <- f:
	}
}

func (m *MainLoop) HandleConsensusMessage(ctx context.Context, message *interfaces.ConsensusRawMessage) {

	select {
	case <-ctx.Done():
		m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "HandleConsensusRawMessage() ID=%s CONTEXT TERMINATED", termincommittee.Str(m.config.Membership.MyMemberId()))
		return

	case m.messagesChannel <- message:
	}

	//	m.worker.HandleConsensusMessage(ctx, message)
}

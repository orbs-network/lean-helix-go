package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/electiontrigger"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/pkg/errors"
	"math"
)

type MainLoop struct {
	messagesChannel        chan *interfaces.ConsensusRawMessage
	mainUpdateStateChannel chan *blockWithProof
	electionTrigger        interfaces.ElectionTrigger
	currentHeight          primitives.BlockHeight
	config                 *interfaces.Config
	logger                 L.LHLogger
	onCommitCallback       interfaces.OnCommitCallback

	worker *WorkerLoop
}

func NewLeanHelix(config *interfaces.Config, onCommitCallback interfaces.OnCommitCallback) *MainLoop {

	var electionTrigger interfaces.ElectionTrigger

	if config.OverrideElectionTrigger != nil {
		electionTrigger = config.OverrideElectionTrigger
	} else {
		electionTrigger = electiontrigger.NewTimerBasedElectionTrigger(config.ElectionTimeoutOnV0, config.OnElectionCB)
	}

	// TODO Create shared State object

	return &MainLoop{
		config:                 config,
		onCommitCallback:       onCommitCallback,
		messagesChannel:        make(chan *interfaces.ConsensusRawMessage, 10), // TODO use config.MsgChanBufLen
		mainUpdateStateChannel: make(chan *blockWithProof, 10),                 // TODO use config.UpdateStateChanBufLen
		electionTrigger:        electionTrigger,
		currentHeight:          0,
		logger:                 LoggerToLHLogger(config.Logger),
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

func (m *MainLoop) RunWorkerLoop(ctx context.Context) {
	lhLog := LoggerToLHLogger(m.config.Logger)
	m.worker = NewWorkerLoop(m.config, lhLog, m.electionTrigger, m.onCommitCallback)
	go m.worker.Run(ctx)
}

func (m *MainLoop) run(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("MAINLOOP PANIC: %v\n", e) // keep this raw print - can be useful if everything breaks
			m.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, m.config.Membership.MyMemberId()), "MAINLOOP PANIC: %v", e)
		}
	}()

	if m.electionTrigger == nil {
		panic("Election trigger was not configured, cannot run Lean Helix (mainloop.run)")
	}

	m.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, m.config.Membership.MyMemberId()), "LHFLOW MAINLOOP START")
	m.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, m.config.Membership.MyMemberId()), "LHMSG MAINLOOP START LISTENING NOW")
	workerCtx, cancelWorkerContext := context.WithCancel(ctx)
	for {
		m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "LHFLOW MAINLOOP LISTENING")
		select {
		case <-ctx.Done(): // system shutdown
			m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "LHFLOW MAINLOOP DONE, Terminating Run().")
			m.logger.Info(L.LC(math.MaxUint64, math.MaxUint64, m.config.Membership.MyMemberId()), "LHMSG MAINLOOP STOPPED LISTENING")
			return

		case message := <-m.messagesChannel:
			parsedMessage := interfaces.ToConsensusMessage(message)
			m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "LHMSG MAINLOOP RECEIVED %v from %v for H=%d", parsedMessage.MessageType(), parsedMessage.SenderMemberId(), parsedMessage.BlockHeight())
			m.worker.MessagesChannel <- &MessageWithContext{ctx: workerCtx, msg: message}

		case trigger := <-m.electionTrigger.ElectionChannel():
			cancelWorkerContext()
			workerCtx, cancelWorkerContext = context.WithCancel(ctx)
			m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "LHFLOW ELECTION - CANCELED WORKER CONTEXT")
			m.worker.electionChannel <- trigger

		case receivedBlockWithProof := <-m.mainUpdateStateChannel: // NodeSync
			if err := checkReceivedBlockIsValid(m.currentHeight, receivedBlockWithProof); err != nil {
				m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "LHFLOW UPDATESTATE - BLOCK IGNORED - %s", err)
				return
			}

			cancelWorkerContext()
			workerCtx, cancelWorkerContext = context.WithCancel(ctx)
			m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "LHFLOW UPDATESTATE - CANCELED WORKER CONTEXT")
			m.worker.workerUpdateStateChannel <- receivedBlockWithProof

		}
	}
}

func checkReceivedBlockIsValid(currentHeight primitives.BlockHeight, receivedBlockWithProof *blockWithProof) error {
	if receivedBlockWithProof == nil {
		return errors.Errorf("receivedBlockWithProof is nil")
	}
	var receivedBlockHeight primitives.BlockHeight
	if receivedBlockWithProof.block == nil {
		receivedBlockHeight = 0
	} else {
		receivedBlockHeight = receivedBlockWithProof.block.Height()
	}
	if receivedBlockHeight < currentHeight {
		return errors.Errorf("Received block height is %d which is lower than current height of %d", receivedBlockWithProof.block.Height(), currentHeight)
	}
	return nil
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
	case m.mainUpdateStateChannel <- &blockWithProof{
		block:               prevBlock,
		prevBlockProofBytes: prevBlockProofBytes,
	}:
	}
}

// called by tests
func (m *MainLoop) TriggerElection(ctx context.Context, f func(ctx context.Context)) {
	if m.electionTrigger == nil {
		panic("Election trigger was not configured, cannot run Lean Helix (mainloop.TriggerElection)")
	}
	select {
	case <-ctx.Done():
		m.logger.Debug(L.LC(m.currentHeight, math.MaxUint64, m.config.Membership.MyMemberId()), "TriggerElection() ID=%s CONTEXT TERMINATED", termincommittee.Str(m.config.Membership.MyMemberId()))
		return
	case m.electionTrigger.ElectionChannel() <- f:
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

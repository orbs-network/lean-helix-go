package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/lean-helix-go/services/electiontrigger"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/pkg/errors"
)

type MainLoop struct {
	messagesChannel             chan *interfaces.ConsensusRawMessage
	mainUpdateStateChannel      chan *blockWithProof
	electionScheduler           interfaces.ElectionScheduler
	config                      *interfaces.Config
	logger                      L.LHLogger
	onCommitCallback            interfaces.OnCommitCallback
	onNewConsensusRoundCallback interfaces.OnNewConsensusRoundCallback
	state                       state.State
	worker                      *WorkerLoop
}

func NewLeanHelix(config *interfaces.Config, onCommitCallback interfaces.OnCommitCallback, onNewConsensusRoundCallback interfaces.OnNewConsensusRoundCallback) *MainLoop {

	var electionTrigger interfaces.ElectionScheduler

	if config.OverrideElectionTrigger != nil {
		electionTrigger = config.OverrideElectionTrigger
	} else {
		electionTrigger = Electiontrigger.NewTimerBasedElectionTrigger(config.ElectionTimeoutOnV0, config.OnElectionCB)
	}

	state := state.NewState()

	return &MainLoop{
		config:                      config,
		onCommitCallback:            onCommitCallback,
		onNewConsensusRoundCallback: onNewConsensusRoundCallback,
		messagesChannel:             make(chan *interfaces.ConsensusRawMessage, 10), // TODO use config.MsgChanBufLen
		mainUpdateStateChannel:      make(chan *blockWithProof, 10),                 // TODO use config.UpdateStateChanBufLen
		electionScheduler:           electionTrigger,
		state:                       state,
		logger:                      L.NewLhLogger(config, state),
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
	m.worker = NewWorkerLoop(
		m.state,
		m.config,
		m.logger,
		m.electionScheduler,
		m.onCommitCallback,
		m.onNewConsensusRoundCallback)
	go m.worker.Run(ctx)
}

func (m *MainLoop) run(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("MAINLOOP PANIC: %v\n", e) // keep this raw print - can be useful if everything breaks
			m.logger.Info("MAINLOOP PANIC: %v", e)
		}
	}()

	if m.electionScheduler == nil {
		panic("Election trigger was not configured, cannot run Lean Helix (mainloop.run)")
	}

	m.logger.Info("LHFLOW LHMSG MAINLOOP START LISTENING NOW")
	workerCtx, cancelWorkerContext := context.WithCancel(ctx)
	for {
		select {
		case <-ctx.Done(): // system shutdown
			m.logger.Info("LHFLOW LHMSG MAINLOOP DONE STOPPED LISTENING, Terminating Run().")
			return

		case message := <-m.messagesChannel:
			parsedMessage := interfaces.ToConsensusMessage(message)
			m.logger.Debug("LHFLOW LHMSG MAINLOOP RECEIVED %v from %v for H=%d", parsedMessage.MessageType(), parsedMessage.SenderMemberId(), parsedMessage.BlockHeight())
			m.worker.MessagesChannel <- &MessageWithContext{ctx: workerCtx, msg: message}

		case trigger := <-m.electionScheduler.ElectionChannel():

			current := m.state.HeightView()
			if current.Height() != trigger.Hv.Height() || current.View() != trigger.Hv.View() { // stale election message
				m.logger.Debug("LHFLOW MAINLOOP ELECTION - INVALID HEIGHT/VIEW IGNORED - Current: %s, ElectionTrigger: %s", current, trigger.Hv)
				continue
			}

			cancelWorkerContext()
			workerCtx, cancelWorkerContext = context.WithCancel(ctx)
			m.logger.Debug("LHFLOW MAINLOOP ELECTION - CANCELED WORKER CONTEXT")
			m.worker.electionChannel <- trigger

		case receivedBlockWithProof := <-m.mainUpdateStateChannel: // NodeSync

			if err := checkReceivedBlockIsValid(m.state.Height(), receivedBlockWithProof); err != nil {
				m.logger.Debug("LHFLOW UPDATESTATE - INVALID MAINLOOP BLOCK IGNORED - %s", err)
				continue
			}

			cancelWorkerContext()
			workerCtx, cancelWorkerContext = context.WithCancel(ctx)
			m.logger.Debug("LHFLOW MAINLOOP UPDATESTATE - CANCELED WORKER CONTEXT")
			m.worker.workerUpdateStateChannel <- receivedBlockWithProof
			m.logger.Debug("LHFLOW MAINLOOP UPDATESTATE - Wrote to worker UpdateState channel")

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
		return errors.Errorf("Received block height is %d which is lower than current height of %d", receivedBlockHeight, currentHeight)
	}
	return nil
}

func (m *MainLoop) ValidateBlockConsensus(ctx context.Context, block interfaces.Block, blockProofBytes []byte, maybePrevBlockProofBytes []byte) error {
	if m.worker == nil {
		panic("ValidateBlockConsensus() worker is nil")
	}
	return m.worker.ValidateBlockConsensus(ctx, block, blockProofBytes, maybePrevBlockProofBytes)
}

// Called from outside to indicate Node Sync
func (m *MainLoop) UpdateState(ctx context.Context, prevBlock interfaces.Block, prevBlockProofBytes []byte) error {

	select {
	case <-ctx.Done():
		m.logger.Debug("UpdateState() ID=%s CONTEXT CANCELED", termincommittee.Str(m.config.Membership.MyMemberId()))
		return errors.Errorf("context canceled")
	case m.mainUpdateStateChannel <- &blockWithProof{
		block:               prevBlock,
		prevBlockProofBytes: prevBlockProofBytes,
	}:
		height := m.state.Height()
		m.logger.Debug("UpdateState() WROTE TO UPDATESTATE MAINLOOP: Block=%v H=%d", prevBlock, height)
		return nil
	}
}

func (m *MainLoop) HandleConsensusMessage(ctx context.Context, message *interfaces.ConsensusRawMessage) {

	select {
	case <-ctx.Done():
		m.logger.Debug("HandleConsensusRawMessage() ID=%s CONTEXT CANCELED", termincommittee.Str(m.config.Membership.MyMemberId()))
		return

	case m.messagesChannel <- message:
	}

	//	m.worker.HandleConsensusMessage(ctx, message)
}

func (m *MainLoop) State() state.State {
	return m.state
}

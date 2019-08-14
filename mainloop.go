package leanhelix

import (
	"context"
	"fmt"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/lean-helix-go/services/electiontrigger"
	"github.com/orbs-network/lean-helix-go/services/interfaces"
	L "github.com/orbs-network/lean-helix-go/services/logger"
	"github.com/orbs-network/lean-helix-go/services/termincommittee"
	"github.com/orbs-network/lean-helix-go/spec/types/go/primitives"
	"github.com/orbs-network/lean-helix-go/spec/types/go/protocol"
	"github.com/orbs-network/lean-helix-go/state"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
	"runtime/debug"
)

type MainLoop struct {
	govnr.TreeSupervisor
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

type govnrErrorer struct {
	logger log.Logger
}

func (h *govnrErrorer) Error(err error) {
	h.logger.Error("recovered panic", log.Error(err), log.String("panic", "true"), log.String("stack-trace", string(debug.Stack())))
}

func GovnrErrorer(logger log.Logger) govnr.Errorer {
	return &govnrErrorer{logger}
}

// TODO Pass logger from Orbs
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
		messagesChannel:             make(chan *interfaces.ConsensusRawMessage), // TODO use config.MsgChanBufLen
		mainUpdateStateChannel:      make(chan *blockWithProof),                 // TODO use config.UpdateStateChanBufLen
		electionScheduler:           electionTrigger,
		state:                       state,
		logger:                      L.NewLhLogger(config, state),
	}
}

// ORBS: LeanHelix.Run(ctx, goroutineLauncher func(f func()) { GoForever(f) }))
// LH: goroutineLauncher(func (){m.runWorkerLoop(ctx)})

type stdoutErrorer struct {
}

func (s stdoutErrorer) Error(err error) {
	fmt.Printf("%s\n", err)
}

func (m *MainLoop) Run(ctx context.Context) {

	m.Supervise(m.runMainLoop(ctx))

	m.Supervise(m.runWorkerLoop(ctx))
}

func (m *MainLoop) runMainLoop(ctx context.Context) *govnr.ForeverHandle {
	logger := log.GetLogger().WithTags(log.Node(m.config.InstanceId.String()), log.String("event_loop", "LHMain"))
	return govnr.Forever(ctx, "lh-mainloop", GovnrErrorer(logger), func() {
		m.run(ctx)
	})
}

// TODO Pass logger from mainloop
func (m *MainLoop) runWorkerLoop(ctx context.Context) govnr.ShutdownWaiter {
	m.worker = NewWorkerLoop(
		m.state,
		m.config,
		m.logger,
		m.electionScheduler,
		m.onCommitCallback,
		m.onNewConsensusRoundCallback)

	logger := log.GetLogger().WithTags(log.Node(m.config.InstanceId.String()), log.String("event_loop", "LHWorker"))
	return govnr.Forever(ctx, "lh-workerloop", GovnrErrorer(logger), func() {
		m.worker.Run(ctx)
	})
}

func (m *MainLoop) run(ctx context.Context) {
	if m.electionScheduler == nil {
		panic("Election trigger was not configured, cannot run Lean Helix (mainloop.run)")
	}

	m.logger.Info("LHFLOW LHMSG MAINLOOP START LISTENING NOW")
	workerCtx, cancelWorkerContext := context.WithCancel(ctx)
	defer cancelWorkerContext()
	var lastSyncedHeight primitives.BlockHeight
	for {
		select {
		case <-ctx.Done(): // system shutdown
			m.logger.Info("LHFLOW LHMSG MAINLOOP DONE STOPPED LISTENING, Terminating Run().")
			return
		case message := <-m.messagesChannel:
			parsedMessage := interfaces.ToConsensusMessage(message)

			m.logger.Debug("LHFLOW LHMSG MAINLOOP RECEIVED %v from %v for H=%d V=%d", parsedMessage.MessageType(), parsedMessage.SenderMemberId(), parsedMessage.BlockHeight(), parsedMessage.View())
			select {
			default: // never block the main loop
			case <-ctx.Done(): // here for uniformity, made redundant by default:
			case m.worker.MessagesChannel <- &MessageWithContext{ctx: workerCtx, msg: message}:
			}

		case trigger := <-m.electionScheduler.ElectionChannel():

			current := m.state.HeightView()
			if current.Height() != trigger.Hv.Height() || current.View() != trigger.Hv.View() { // stale election message
				m.logger.Debug("LHFLOW ELECTION MAINLOOP - INVALID HEIGHT/VIEW IGNORED - Current: %s, ElectionTrigger: %s", current, trigger.Hv)
				continue
			}

			cancelWorkerContext()
			workerCtx, cancelWorkerContext = context.WithCancel(ctx)
			m.logger.Debug("LHFLOW ELECTION MAINLOOP - CANCELED WORKER CONTEXT (received election trigger with H=%d V=%d)", trigger.Hv.Height(), trigger.Hv.View())
			message := &workerElectionsTriggerMessage{
				ctx:             workerCtx,
				ElectionTrigger: trigger,
			}
			select {
			case <-ctx.Done(): // system shutdown
			case m.worker.electionChannel <- message:
			}

		case receivedBlockWithProof := <-m.mainUpdateStateChannel: // NodeSync

			effectiveLastHeight := m.state.Height()
			if effectiveLastHeight < lastSyncedHeight {
				effectiveLastHeight = lastSyncedHeight
			}
			if err := checkReceivedBlockIsValid(effectiveLastHeight, receivedBlockWithProof); err != nil {
				m.logger.Debug("LHFLOW UPDATESTATE MAINLOOP - INVALID BLOCK IGNORED - %s", err)
				continue
			}

			cancelWorkerContext()
			workerCtx, cancelWorkerContext = context.WithCancel(ctx)
			m.logger.Debug("LHFLOW UPDATESTATE MAINLOOP - CANCELED WORKER CONTEXT")
			message := &workerUpdateStateMessage{
				ctx:            workerCtx,
				blockWithProof: receivedBlockWithProof,
			}
			select {
			case <-ctx.Done(): // system shutdown
			case m.worker.workerUpdateStateChannel <- message:
			}

			if receivedBlockWithProof.block != nil {
				lastSyncedHeight = receivedBlockWithProof.block.Height()
			}

			m.logger.Debug("LHFLOW UPDATESTATE MAINLOOP - Wrote to worker UpdateState channel")
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

// Used by orbs-network-go
func GetMemberIdsFromBlockProof(blockProofBytes []byte) ([]primitives.MemberId, error) {
	if blockProofBytes == nil || len(blockProofBytes) == 0 {
		return nil, errors.Errorf("GetMemberIdsFromBlockProof: nil blockProof - cannot deduce members locally")
	}
	blockProof := protocol.BlockProofReader(blockProofBytes)
	sendersIterator := blockProof.NodesIterator()
	committeeMembers := make([]primitives.MemberId, 0)
	for sendersIterator.HasNext() {
		committeeMembers = append(committeeMembers, sendersIterator.NextNodes().MemberId())
	}
	return committeeMembers, nil
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

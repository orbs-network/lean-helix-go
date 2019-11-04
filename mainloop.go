package leanhelix

import (
	"context"
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
	"time"
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
	state                       *state.State
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
		messagesChannel:             make(chan *interfaces.ConsensusRawMessage),
		mainUpdateStateChannel:      make(chan *blockWithProof),
		electionScheduler:           electionTrigger,
		state:                       state,
		logger:                      L.NewLhLogger(config, state),
	}
}

func (m *MainLoop) Run(ctx context.Context) govnr.ShutdownWaiter {

	startTime := time.Now()
	m.worker = NewWorkerLoop(
		m.state,
		m.config,
		m.logger,
		m.electionScheduler,
		m.onCommitCallback,
		m.onNewConsensusRoundCallback)

	m.Supervise(m.runMainLoop(ctx))

	logger := log.GetLogger().WithTags(log.Node(m.config.InstanceId.String()), log.String("event_loop", "LHWorker"))
	m.Supervise(govnr.Forever(ctx, "lh-workerloop", GovnrErrorer(logger), func() {
		m.worker.Run(ctx)
	}))

	m.logger.Info("MainLoop.Run() completed in %d ms", (time.Now().Sub(startTime))/1000000)
	return m

}

func (m *MainLoop) runMainLoop(ctx context.Context) *govnr.ForeverHandle {
	logger := log.GetLogger().WithTags(log.Node(m.config.InstanceId.String()), log.String("event_loop", "LHMain"))
	return govnr.Forever(ctx, "lh-mainloop", GovnrErrorer(logger), func() {
		m.run(ctx)
	})
}

func (m *MainLoop) run(ctx context.Context) {
	if m.electionScheduler == nil {
		panic("Election trigger was not configured, cannot run Lean Helix (mainloop.run)")
	}

	m.logger.Info("LHFLOW LHMSG MAINLOOP START LISTENING NOW")

	var maxBlockHeightBySync *primitives.BlockHeight
	var shutdown bool
	for !shutdown {
		m.state.GcOldContexts()
		select {
		case <-ctx.Done(): // system shutdown
			shutdown = true

		case message := <-m.messagesChannel:
			parsedMessage := interfaces.ToConsensusMessage(message)

			m.logger.Debug("LHFLOW LHMSG MAINLOOP RECEIVED %v from %v for H=%d V=%d", parsedMessage.MessageType(), parsedMessage.SenderMemberId(), parsedMessage.BlockHeight(), parsedMessage.View())

			_, err := m.state.Contexts.For(state.NewHeightView(parsedMessage.BlockHeight(), parsedMessage.View()))
			if err != nil {
				m.logger.Debug("LHFLOW LHMSG MAINLOOP - IGNORING RECEIVED MESSAGE %v FROM %v WITH %e", parsedMessage.MessageType(), parsedMessage.SenderMemberId(), err)
				continue
			}

			select {
			default: // never block the main loop
			case <-ctx.Done(): // here for uniformity, made redundant by default:
			case m.worker.MessagesChannel <- message:
			}

		case trigger := <-m.electionScheduler.ElectionChannel():
			targetHv := state.NewHeightView(trigger.Hv.Height(), trigger.Hv.View()+1)
			m.state.Contexts.CancelOlderThan(targetHv)
			_, err := m.state.Contexts.For(targetHv)
			if err != nil {
				m.logger.Debug("LHFLOW LHMSG MAINLOOP - IGNORING ELECTION TRIGGER WITH %e", err)
				continue
			}

			m.logger.Debug("LHFLOW ELECTION MAINLOOP - CANCELED WORKER CONTEXT (received election trigger with H=%d V=%d)", trigger.Hv.Height(), trigger.Hv.View())
			m.sendElectionMessageNonBlocking(ctx, trigger)

		case receivedBlockWithProof := <-m.mainUpdateStateChannel: // NodeSync
			if receivedBlockWithProof == nil {
				m.logger.Debug("LHFLOW UPDATESTATE MAINLOOP - INVALID BLOCK IGNORED - receivedBlockWithProof is nil")
				continue
			}
			var receivedBlockHeight primitives.BlockHeight
			if receivedBlockWithProof.block == nil {
				receivedBlockHeight = 0
			} else {
				receivedBlockHeight = receivedBlockWithProof.block.Height()
			}

			if maxBlockHeightBySync != nil && *maxBlockHeightBySync >= receivedBlockHeight {
				m.logger.Debug("LHFLOW UPDATESTATE MAINLOOP - Already received a more recent update message than block %d", receivedBlockHeight)
				continue
			}

			hv := state.NewHeightView(receivedBlockHeight+1, 0)
			m.state.Contexts.CancelOlderThan(hv)

			_, err := m.state.Contexts.For(hv)
			if err != nil {
				m.logger.Debug("LHFLOW LHMSG MAINLOOP - IGNORING BLOCK SYNC WITH %e", err)
				continue
			}

			m.logger.Debug("LHFLOW UPDATESTATE MAINLOOP - CANCELED WORKER CONTEXT")
			message := receivedBlockWithProof

			err = m.sendUpdateMessageNonBlocking(ctx, message)
			if err != nil {
				continue
			}

			if maxBlockHeightBySync == nil {
				maxBlockHeightBySync = new(primitives.BlockHeight)
			}
			*maxBlockHeightBySync = receivedBlockHeight
			m.logger.Debug("LHFLOW UPDATESTATE MAINLOOP - Wrote to worker UpdateState channel")
		}
	}

	m.logger.Info("LHFLOW LHMSG MAINLOOP DONE STOPPED LISTENING, SHUTDOWN END")
	m.worker.interrupt()
}

func (m *MainLoop) sendElectionMessageNonBlocking(ctx context.Context, trigger *interfaces.ElectionTrigger) {
	elChannel :=  m.worker.electionChannel
	bufferSize := cap(elChannel)
	if bufferSize == 0 {
		panic("electionChannel buffer size must be at least 1")
	}

	if len(elChannel) == bufferSize { // full buffer
		select {
		case <-elChannel: // free one slot
		default: // worker raced us and emptied buffer
		}
	}

	select {
	case <-ctx.Done(): // system shutdown
	case elChannel <- trigger:
	}
}

func (m *MainLoop) sendUpdateMessageNonBlocking(ctx context.Context, blockWithProof *blockWithProof) error {
	msgChannel :=  m.worker.workerUpdateStateChannel
	bufferSize := cap(msgChannel)
	if bufferSize == 0 {
		panic("workerUpdateStateChannel buffer size must be at least 1")
	}

	if len(msgChannel) == bufferSize { // full buffer
		select {
		case <-msgChannel: // free one slot
		default: // worker raced us and emptied buffer
		}
	}

	select {
	case <-ctx.Done(): // system shutdown
		return ctx.Err()
	case msgChannel <- blockWithProof:
		return nil
	}
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

func (m *MainLoop) State() *state.State {
	return m.state
}

package poc_2thr

import (
	"context"
	"fmt"
	"github.com/go-playground/ansi"
	"runtime"
	"sync"
	"time"
)

type MessageType int

const PREPREPARE MessageType = 1
const PREPARE MessageType = 2
const COMMIT MessageType = 3
const VIEW_CHANGE MessageType = 4
const NEW_VIEW MessageType = 5

type Message struct {
	msgType MessageType
	block   *Block
	view    int
}

func (m *Message) String() string {
	return fmt.Sprintf("<<Type=%d H=%d V=%d>>", m.msgType, m.block.h, m.view)
}

type Block struct {
	h int
}

type Config struct {
	CancelTestAfter      time.Duration
	WaitAfterCancelTest  time.Duration
	CreateBlock          time.Duration
	ValidateBlock        time.Duration
	MessageChannelBufLen int
}

// Using WaitGroups
// See https://nathanleclaire.com/blog/2014/02/15/how-to-wait-for-all-goroutines-to-finish-executing-before-continuing/
func Run(config *Config) {
	timeToCancel := config.CancelTestAfter
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), "ID", "ROOT"))
	Log("Run() start timeToCancel=%s starting *ORBS* goroutine", timeToCancel)
	var wg sync.WaitGroup
	go runOrbs(ctx, &wg, config)
	time.Sleep(timeToCancel)
	Log("Run() CANCELLING TEST ON TIMEOUT")
	cancel()
	time.Sleep(config.WaitAfterCancelTest)
	numGoroutineAfter := runtime.NumGoroutine()
	Log("Goroutines=%config WAIT FOR WG", numGoroutineAfter)
	wg.Wait()
	Log("********** Run() end. Goroutines=%config", numGoroutineAfter)
}

func NewBlock(h int) *Block {
	return &Block{
		h: h,
	}
}

func Log(format string, a ...interface{}) {
	fmt.Printf(Yellow+time.Now().Format("03:04:05.000")+" "+format+"\n", a...)
}

func NewPPM(block *Block) *Message {
	return &Message{
		msgType: PREPREPARE,
		block:   block,
	}
}

func NewCM(block *Block) *Message {
	return &Message{
		msgType: COMMIT,
		block:   block,
	}
}

// Adapted from github.com/orbs-network/scribe/log/formatters.go
func colorize(lineId string) string {
	colors := []string{ansi.Cyan, ansi.Yellow, ansi.LightBlue, ansi.Magenta, ansi.LightYellow, ansi.LightRed, ansi.LightGreen, ansi.LightMagenta, ansi.Green}
	fourthBeforeLastChar := int(lineId[len(lineId)-4])
	return colors[fourthBeforeLastChar%len(colors)]
}

// Copied from github.com/go-playground/ansi/ansi.go
// EscSeq is a predefined ANSI escape sequence
type EscSeq string

// ANSI escape sequences
// NOTE: in a standard xterm terminal the light colors will appear BOLD instead of the light variant
const (
	Reset             EscSeq = "\x1b[0m"
	Italics                  = "\x1b[3m"
	Underline                = "\x1b[4m"
	Blink                    = "\x1b[5m"
	Inverse                  = "\x1b[7m"
	ItalicsOff               = "\x1b[23m"
	UnderlineOff             = "\x1b[24m"
	BlinkOff                 = "\x1b[25m"
	InverseOff               = "\x1b[27m"
	Black                    = "\x1b[30m"
	DarkGray                 = "\x1b[30;1m"
	Red                      = "\x1b[31m"
	LightRed                 = "\x1b[31;1m"
	Green                    = "\x1b[32m"
	LightGreen               = "\x1b[32;1m"
	Yellow                   = "\x1b[33m"
	LightYellow              = "\x1b[33;1m"
	Blue                     = "\x1b[34m"
	LightBlue                = "\x1b[34;1m"
	Magenta                  = "\x1b[35m"
	LightMagenta             = "\x1b[35;1m"
	Cyan                     = "\x1b[36m"
	LightCyan                = "\x1b[36;1m"
	Gray                     = "\x1b[37m"
	White                    = "\x1b[37;1m"
	ResetForeground          = "\x1b[39m"
	BlackBackground          = "\x1b[40m"
	RedBackground            = "\x1b[41m"
	GreenBackground          = "\x1b[42m"
	YellowBackground         = "\x1b[43m"
	BlueBackground           = "\x1b[44m"
	MagentaBackground        = "\x1b[45m"
	CyanBackground           = "\x1b[46m"
	GrayBackground           = "\x1b[47m"
	ResetBackground          = "\x1b[49m"
	Bold                     = "\x1b[1m"
	BoldOff                  = "\x1b[22m"
)

// Left out due to not being widely supported:
// StrikethroughOn           = "\x1b[9m"
// StrikethroughOff          = "\x1b[29m"

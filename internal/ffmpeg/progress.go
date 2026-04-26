package ffmpeg

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type State struct {
	FPS     float64
	Bitrate string
	OutTime time.Duration
	Speed   string
}

type Parser struct {
	mu       sync.Mutex
	live     State
	pending  State
	onUpdate func(State)
}

func NewParser(onUpdate func(State)) *Parser {
	return &Parser{onUpdate: onUpdate}
}

// Feed is compatible with ezec's CallbackConsumer — called once per line from pipe:3.
func (p *Parser) Feed(line string) {
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return
	}

	switch key {
	case "fps":
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			p.pending.FPS = f
		}
	case "bitrate":
		p.pending.Bitrate = value
	case "out_time_us":
		if us, err := strconv.ParseInt(value, 10, 64); err == nil {
			p.pending.OutTime = time.Duration(us) * time.Microsecond
		}
	case "speed":
		p.pending.Speed = value
	case "progress":
		p.mu.Lock()
		p.live = p.pending
		snapshot := p.live
		p.mu.Unlock()
		p.pending = State{}
		p.onUpdate(snapshot)
	}
}

// State returns the most recently completed progress state. Safe to call
// concurrently with Feed.
func (p *Parser) State() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.live
}

// Render formats a State into the lilcaster status line.
func Render(scenarioName string, s State) string {
	total := int(s.OutTime.Seconds())
	h := total / 3600
	m := (total % 3600) / 60
	sec := total % 60
	return fmt.Sprintf("lilcaster: [%s] %02d:%02d:%02d | fps=%.1f | bitrate=%s | speed=%s",
		scenarioName, h, m, sec, s.FPS, s.Bitrate, s.Speed)
}

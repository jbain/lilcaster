package ffmpeg

import (
	"testing"
	"time"
)

func TestParser_ContinueBlock(t *testing.T) {
	var calls int
	var last State
	p := NewParser(func(s State) {
		calls++
		last = s
	})

	lines := []string{
		"fps=30.5",
		"bitrate=2500.0kbits/s",
		"out_time_us=83000000",
		"speed=1.0x",
		"progress=continue",
	}
	for _, l := range lines {
		p.Feed(l)
	}

	if calls != 1 {
		t.Errorf("callback count: got %d, want 1", calls)
	}
	if last.FPS != 30.5 {
		t.Errorf("FPS: got %v, want 30.5", last.FPS)
	}
	if last.Bitrate != "2500.0kbits/s" {
		t.Errorf("Bitrate: got %q, want %q", last.Bitrate, "2500.0kbits/s")
	}
	want := 83 * time.Second
	if last.OutTime != want {
		t.Errorf("OutTime: got %v, want %v", last.OutTime, want)
	}
	if last.Speed != "1.0x" {
		t.Errorf("Speed: got %q, want %q", last.Speed, "1.0x")
	}

	// State() should reflect the completed block.
	s := p.State()
	if s.FPS != 30.5 {
		t.Errorf("State().FPS: got %v", s.FPS)
	}
}

func TestParser_EndBlock(t *testing.T) {
	var calls int
	p := NewParser(func(State) { calls++ })

	p.Feed("fps=25.0")
	p.Feed("progress=end")

	if calls != 1 {
		t.Errorf("callback count: got %d, want 1", calls)
	}
	if p.State().FPS != 25.0 {
		t.Errorf("FPS after end: got %v", p.State().FPS)
	}
}

func TestParser_TwoBlocks(t *testing.T) {
	var calls int
	p := NewParser(func(State) { calls++ })

	for i := 0; i < 2; i++ {
		p.Feed("fps=30.0")
		p.Feed("progress=continue")
	}

	if calls != 2 {
		t.Errorf("callback count: got %d, want 2", calls)
	}
}

func TestParser_PendingStateResetBetweenBlocks(t *testing.T) {
	p := NewParser(func(State) {})

	p.Feed("fps=30.0")
	p.Feed("bitrate=1000kbits/s")
	p.Feed("progress=continue")

	// Second block only sets fps — bitrate should be gone.
	p.Feed("fps=29.0")
	p.Feed("progress=continue")

	s := p.State()
	if s.FPS != 29.0 {
		t.Errorf("FPS: got %v, want 29.0", s.FPS)
	}
	if s.Bitrate != "" {
		t.Errorf("Bitrate should be empty after reset, got %q", s.Bitrate)
	}
}

func TestParser_UnknownKeysIgnored(t *testing.T) {
	var calls int
	p := NewParser(func(State) { calls++ })

	p.Feed("unknown_key=whatever")
	p.Feed("fps=10.0")
	p.Feed("progress=continue")

	if calls != 1 {
		t.Errorf("callback count: got %d, want 1", calls)
	}
	if p.State().FPS != 10.0 {
		t.Errorf("FPS: got %v", p.State().FPS)
	}
}

func TestRender(t *testing.T) {
	s := State{
		FPS:     30.0,
		Bitrate: "2500.0kbits/s",
		OutTime: 83*time.Second + 500*time.Millisecond,
		Speed:   "1.0x",
	}
	got := Render("simple_stream", s)
	want := "lilcaster: [simple_stream] 00:01:23 | fps=30.0 | bitrate=2500.0kbits/s | speed=1.0x"
	if got != want {
		t.Errorf("Render():\n got  %q\n want %q", got, want)
	}
}

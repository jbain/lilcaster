package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/jbain/ezec"
	"github.com/jbain/ezec/pkg/consumers"
)

type Process struct {
	Cmd    *ezec.Cmd
	Parser *Parser

	progressConsumer *consumers.CallbackConsumer
	stderrRing       *consumers.LastNConsumer
	stderrVerbose    *consumers.CallbackConsumer // nil in non-verbose mode
}

func New(args SliceArgs, parser *Parser, verbose bool) (*Process, error) {
	binary, err := findFFmpeg()
	if err != nil {
		return nil, err
	}

	cmd := ezec.Command(binary, args)

	progressConsumer := consumers.NewCallbackConsumer("progress", 64, parser.Feed)
	cmd.AddFd([]ezec.LineConsumer{progressConsumer})

	stderrRing := consumers.NewLastNConsumer("stderr-ring", 64, 10)
	stderrConsumers := []ezec.LineConsumer{stderrRing}

	var stderrVerbose *consumers.CallbackConsumer
	if verbose {
		stderrVerbose = consumers.NewCallbackConsumer("stderr-verbose", 64, func(line string) {
			fmt.Fprintln(os.Stderr, line)
		})
		stderrConsumers = append(stderrConsumers, stderrVerbose)
	}
	cmd.Stderr = stderrConsumers

	return &Process{
		Cmd:              cmd,
		Parser:           parser,
		progressConsumer: progressConsumer,
		stderrRing:       stderrRing,
		stderrVerbose:    stderrVerbose,
	}, nil
}

func (p *Process) Start() error {
	go p.progressConsumer.Start()
	go p.stderrRing.Start()
	if p.stderrVerbose != nil {
		go p.stderrVerbose.Start()
	}
	return p.Cmd.Start()
}

// StderrLines returns the last (up to 10) lines written to ffmpeg's stderr.
// Safe to call after Wait returns.
func (p *Process) StderrLines() []string {
	return p.stderrRing.Lines()
}

func (p *Process) Wait() error {
	return p.Cmd.Wait()
}

// Kill sends SIGTERM and schedules a SIGKILL after 3 seconds if the process
// hasn't exited. The caller's Wait() will unblock once the process is reaped.
func (p *Process) Kill() error {
	if p.Cmd.Process == nil {
		return nil
	}
	if err := p.Cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return p.Cmd.Process.Kill()
	}
	go func() {
		time.Sleep(3 * time.Second)
		_ = p.Cmd.Process.Kill()
	}()
	return nil
}

func findFFmpeg() (string, error) {
	if path := os.Getenv("FFMPEG_PATH"); path != "" {
		return path, nil
	}
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("ffmpeg not found: set FFMPEG_PATH or ensure ffmpeg is in $PATH")
	}
	return path, nil
}

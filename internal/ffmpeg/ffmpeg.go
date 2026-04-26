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
	stderrConsumer   *consumers.CallbackConsumer
}

func New(args SliceArgs, parser *Parser, verbose bool) (*Process, error) {
	binary, err := findFFmpeg()
	if err != nil {
		return nil, err
	}

	cmd := ezec.Command(binary, args)

	progressConsumer := consumers.NewCallbackConsumer("progress", 64, parser.Feed)
	cmd.AddFd([]ezec.LineConsumer{progressConsumer})

	var stderrFn func(string)
	if verbose {
		stderrFn = func(line string) { fmt.Fprintln(os.Stderr, line) }
	} else {
		stderrFn = func(string) {}
	}
	stderrConsumer := consumers.NewCallbackConsumer("stderr", 64, stderrFn)
	cmd.Stderr = []ezec.LineConsumer{stderrConsumer}

	return &Process{
		Cmd:              cmd,
		Parser:           parser,
		progressConsumer: progressConsumer,
		stderrConsumer:   stderrConsumer,
	}, nil
}

func (p *Process) Start() error {
	go p.progressConsumer.Start()
	go p.stderrConsumer.Start()
	return p.Cmd.Start()
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

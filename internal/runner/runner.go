package runner

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lilcaster/internal/config"
	"lilcaster/internal/ffmpeg"
)

type Overrides struct {
	Source string
	Sink   string
}

func Run(sc config.Scenario, ov Overrides, verbose bool) error {
	if ov.Source != "" {
		sc.Sources = []config.Endpoint{{Path: ov.Source}}
	}
	if ov.Sink != "" {
		sc.Sinks = []config.Endpoint{{Path: ov.Sink}}
	}

	sources, err := resolveAll(sc.Sources)
	if err != nil {
		return err
	}
	sinks, err := resolveAll(sc.Sinks)
	if err != nil {
		return err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1)
	defer signal.Stop(sigCh)

	for {
		filterStr := ffmpeg.Build(sc.Filters)
		args := ffmpeg.BuildArgs(sc, sources, sinks, filterStr)

		parser := ffmpeg.NewParser(func(ffmpeg.State) {})
		proc, err := ffmpeg.New(args, parser, verbose)
		if err != nil {
			return err
		}
		if err := proc.Start(); err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		if !verbose {
			go runTicker(ctx, sc.Name, parser)
		}

		waitDone := make(chan error, 1)
		go func() { waitDone <- proc.Wait() }()

		type action int
		const (
			actExit   action = iota
			actStop          // SIGINT / SIGTERM
			actSIGHUP        // re-resolve + restart
			actSIGUSR1       // reuse paths + restart
		)

		var (
			act     action
			waitErr error
		)
		select {
		case waitErr = <-waitDone:
			act = actExit
		case sig := <-sigCh:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				act = actStop
			case syscall.SIGHUP:
				act = actSIGHUP
			case syscall.SIGUSR1:
				act = actSIGUSR1
			}
		}

		cancel()

		if act != actExit {
			_ = proc.Kill()
			<-waitDone
		}

		if !verbose {
			fmt.Println()
		}

		if act == actExit && waitErr != nil {
			fmt.Fprintf(os.Stderr, "ffmpeg exited with error: %v\n", waitErr)
			for _, line := range proc.StderrLines() {
				fmt.Fprintln(os.Stderr, line)
			}
		}

		switch act {
		case actExit, actStop:
			return waitErr

		case actSIGHUP:
			sources, err = reResolveAll(sources)
			if err != nil {
				return err
			}
			sinks, err = reResolveAll(sinks)
			if err != nil {
				return err
			}

		case actSIGUSR1:
			// reuse cached resolved paths — no action needed
		}
	}
}

func runTicker(ctx context.Context, scenarioName string, parser *ffmpeg.Parser) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Printf("\r\033[K%s", ffmpeg.Render(scenarioName, parser.State()))
		case <-ctx.Done():
			return
		}
	}
}

func resolveAll(endpoints []config.Endpoint) ([]ffmpeg.ResolvedEndpoint, error) {
	resolved := make([]ffmpeg.ResolvedEndpoint, len(endpoints))
	for i, ep := range endpoints {
		re, err := ffmpeg.Resolve(ep)
		if err != nil {
			return nil, fmt.Errorf("resolving %q: %w", ep.Path, err)
		}
		resolved[i] = re
	}
	return resolved, nil
}

func reResolveAll(resolved []ffmpeg.ResolvedEndpoint) ([]ffmpeg.ResolvedEndpoint, error) {
	result := make([]ffmpeg.ResolvedEndpoint, len(resolved))
	for i, re := range resolved {
		newRe, err := ffmpeg.Resolve(re.Original)
		if err != nil {
			return nil, fmt.Errorf("re-resolving %q: %w", re.Original.Path, err)
		}
		result[i] = newRe
	}
	return result, nil
}

package ffmpeg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"lilcaster/internal/config"
)

type ResolvedEndpoint struct {
	Original config.Endpoint
	Path     string
}

func Resolve(e config.Endpoint) (ResolvedEndpoint, error) {
	re := ResolvedEndpoint{Original: e}

	switch {
	case strings.HasPrefix(e.Path, "script://"):
		scriptPath := strings.TrimPrefix(e.Path, "script://")
		if !strings.HasPrefix(scriptPath, "/") {
			scriptPath = "./" + scriptPath
		}
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(scriptPath, e.ScriptArgs...)
		cmd.Stdout = &stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
		if err := cmd.Run(); err != nil {
			return re, fmt.Errorf("script %q: %w\nstderr: %s", scriptPath, err, stderr.String())
		}
		re.Path = strings.TrimSpace(stdout.String())

	case strings.HasPrefix(e.Path, "avfoundation://"):
		return re, fmt.Errorf("avfoundation not yet supported")

	default:
		re.Path = e.Path
	}

	return re, nil
}

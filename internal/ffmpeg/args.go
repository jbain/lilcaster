package ffmpeg

import (
	"strconv"

	"lilcaster/internal/config"
)

type SliceArgs []string

func (s SliceArgs) Args() []string { return s }

// BuildArgs produces the full ffmpeg argument list for a scenario.
func BuildArgs(
	sc config.Scenario,
	sources []ResolvedEndpoint,
	sinks []ResolvedEndpoint,
	filterStr string,
) SliceArgs {
	args := SliceArgs{"-progress", "pipe:3"}

	for _, src := range sources {
		if sc.Loop != 0 {
			args = append(args, "-stream_loop", strconv.Itoa(sc.Loop))
		}
		args = append(args, "-re", "-i", src.Path)
	}

	if filterStr != "" {
		args = append(args, "-vf", filterStr)
	}

	args = append(args, "-c:v", "libx264", "-preset", "veryfast", "-c:a", "aac", "-f", "flv")

	for _, sink := range sinks {
		args = append(args, sink.Path)
	}

	return args
}

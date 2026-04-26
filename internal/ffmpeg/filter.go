package ffmpeg

import (
	"strings"

	"lilcaster/internal/config"
)

const timestampFilter = "drawtext=text='%{localtime\\:%T}':x=10:y=10:fontsize=24:fontcolor=white:box=1:boxcolor=black@0.5" +
	",drawtext=text='%{pts\\:hms}':x=10:y=44:fontsize=24:fontcolor=white:box=1:boxcolor=black@0.5"

// Build returns the value for ffmpeg's -vf flag. Returns empty string when
// filters is empty (caller should omit the -vf flag entirely).
func Build(filters []config.FilterEntry) string {
	var parts []string
	for _, entry := range filters {
		switch f := entry.Filter.(type) {
		case *config.ScaleFilter:
			parts = append(parts, "scale="+f.Width+":"+f.Height)
		case *config.TimestampFilter:
			parts = append(parts, timestampFilter)
		case *config.CustomFilter:
			parts = append(parts, f.String)
		}
	}
	return strings.Join(parts, ",")
}

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lilcaster",
	Short: "lilcaster — ffmpeg-wrapped livestream runner",
	Long: `lilcaster is a CLI tool that wraps ffmpeg to simplify running and managing
livestreams. It supports multiple named scenarios defined in a config file, each
with configurable sources, sinks, filters, and loop behavior.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lilcaster/internal/config"
	"lilcaster/internal/runner"
)

var scenarioCmd = &cobra.Command{
	Use:   "scenario <name>",
	Short: "Run a named scenario from the config file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		name := args[0]
		var found *config.Scenario
		var names []string
		for i := range cfg.Scenarios {
			names = append(names, cfg.Scenarios[i].Name)
			if cfg.Scenarios[i].Name == name {
				found = &cfg.Scenarios[i]
			}
		}
		if found == nil {
			return fmt.Errorf("scenario %q not found (available: %s)", name, strings.Join(names, ", "))
		}

		source, _ := cmd.Flags().GetString("source")
		sink, _ := cmd.Flags().GetString("sink")
		verbose, _ := cmd.Flags().GetBool("verbose")

		return runner.Run(*found, runner.Overrides{Source: source, Sink: sink}, verbose)
	},
}

func init() {
	scenarioCmd.Flags().String("source", "", "replace all sources with a single source path")
	scenarioCmd.Flags().String("sink", "", "replace all sinks with a single sink path")
	scenarioCmd.Flags().Bool("verbose", false, "stream ffmpeg stderr output live")
	rootCmd.AddCommand(scenarioCmd)
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toshism/gotestit/watcher"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tests for the provided files",
	Long:  `Accepts a list of files and attempts to locate relevant test files and run them.`,
	Run: func(cmd *cobra.Command, args []string) {
		wg := NewWatchGroup(project)
		watcher.RunTests(wg, args)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

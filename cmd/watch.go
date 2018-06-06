package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/toshism/gotestit/watcher"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch PROJECT for changes and run tests",
	Run: func(cmd *cobra.Command, args []string) {
		watch()
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.PersistentFlags().StringVarP(&project, "project", "p", "", "Project")
}

func NewWatchGroup(project string) watcher.WatchGroup {
	projects := viper.Sub("projects")
	projectConfig := projects.Sub(project)
	var watchGroup watcher.WatchGroup
	projectConfig.Unmarshal(&watchGroup)
	watchGroup.Name = project
	return watchGroup
}

func watch() {
	var ws []watcher.WatchGroup
	if project != "" {
		watchGroup := NewWatchGroup(project)
		ws = append(ws, watchGroup)
	} else {
		for project := range viper.GetStringMap("projects") {
			watchGroup := NewWatchGroup(project)
			ws = append(ws, watchGroup)
		}
	}
	watcher.Watch(ws)
}

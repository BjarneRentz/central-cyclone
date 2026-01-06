package projects

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var syncprojectsCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs projects from DependencyTrack to the local configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Sync projects")
	},
}

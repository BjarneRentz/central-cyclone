package projects

import (
	"central-cyclone/cmd/extensions"
	"central-cyclone/internal/config"
	"central-cyclone/internal/dt"
	"context"
	"log/slog"

	"github.com/spf13/cobra"
)

var syncprojectsCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs projects from DependencyTrack to the local configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		settings, err := extensions.GetSettings(cmd)
		if err != nil {
			slog.Error("Could not get settings from context", "error", err)
			return err
		}

		projects := []config.Project{}

		for _, app := range settings.Applications {
			projects = append(projects, app.Projects...)
		}
		dtClient, err := dt.NewDTrackClient(&settings.DependencyTrack)
		if err != nil {
			slog.Error("Could not create Dependency-Track client", "error", err)
			return err
		}

		projectsSyncer := dt.ProjectSyncer{Client: dtClient}

		projectsSyncer.SyncProjects(context.TODO(), projects)

		return nil
	},
}

func init() {
	extensions.RequireConfig(syncprojectsCmd)
}

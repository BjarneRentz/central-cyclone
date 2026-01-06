package dtrack

import (
	"central-cyclone/cmd/dtrack/projects"

	"github.com/spf13/cobra"
)

var DtCmd = &cobra.Command{
	Use:   "dt",
	Short: "Commands related to DependencyTrack",
}

func init() {
	DtCmd.AddCommand(projects.ProjectsCmd)
}

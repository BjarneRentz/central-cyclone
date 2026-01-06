package projects

import (
	"github.com/spf13/cobra"
)

var ProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Commands related to DependencyTrack projects",
}

func init() {
	ProjectsCmd.AddCommand(syncprojectsCmd)
}

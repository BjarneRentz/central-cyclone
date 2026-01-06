/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"central-cyclone/internal/config"
	coordinator "central-cyclone/internal/handlers"
	"central-cyclone/internal/upload"
	"central-cyclone/internal/workspace"
	"log/slog"

	"github.com/spf13/cobra"
)

var uploadSboms bool

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyzes all configured resources and creates SBOMs",

	RunE: func(cmd *cobra.Command, args []string) error {
		settings, err := GetSettings(cmd)
		if err != nil {
			slog.Error("Could not get settings from context", "error", err)
			return err
		}
		runAnalyzeCommand(settings)
		return nil
	},
}

func init() {
	requireConfig(analyzeCmd)
	analyzeCmd.Flags().BoolVar(&uploadSboms, "upload", false, "Upload SBOMs to DependencyTrack after generation")
}

func runAnalyzeCommand(settings *config.Settings) {

	workspaceHandler, err := workspace.CreateLocalWorkspace()
	if err != nil {
		slog.Error("Error creating workspace", "error", err)
		return
	}
	err = workspaceHandler.Clear()
	if err != nil {
		slog.Error("Error clearing workspace", "error", err)
		return
	}

	if uploadSboms {
		uploader, err := upload.CreateDependencyTrackUploader(settings)
		if err != nil {
			slog.Error("Error creating uploader", "error", err)
			return
		}
		coordinator.AnalyzeAndUpload(settings, workspaceHandler, uploader)
	} else {
		coordinator.AnalyzeAndSave(settings, workspaceHandler)
	}
}

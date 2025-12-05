/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"central-cyclone/internal/config"
	coordinator "central-cyclone/internal/handlers"
	"central-cyclone/internal/upload"
	"central-cyclone/internal/workspace"
	"fmt"

	"github.com/spf13/cobra"
)

var cfgFile string
var uploadSboms bool

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyzes all configured resources and creates SBOMs",

	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.LoadFromFile(cfgFile)
		if err != nil {
			fmt.Printf("Error loading configuration: %v\n", err)
			return
		}
		runAnalyzeCommand(config)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&cfgFile, "config", "c", "./config.json", "Path to the configuration file")
	analyzeCmd.Flags().BoolVar(&uploadSboms, "upload", false, "Upload SBOMs to DependencyTrack after generation")
}

func runAnalyzeCommand(settings *config.Settings) {

	workspaceHandler, err := workspace.CreateLocalWorkspace()
	if err != nil {
		fmt.Printf("Error creating workspace: %v\n", err)
		return
	}
	err = workspaceHandler.Clear()
	if err != nil {
		fmt.Printf("Error clearing workspace: %v\n", err)
		return
	}

	if uploadSboms {
		uploader, err := upload.CreateDependencyTrackUploader(settings)
		if err != nil {
			fmt.Printf("Error creating uploader: %v\n", err)
			return
		}
		coordinator.AnalyzeAndUpload(settings, workspaceHandler, uploader)
	} else {
		coordinator.AnalyzeAndSave(settings, workspaceHandler)
	}
}

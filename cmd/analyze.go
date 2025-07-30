/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/manager"
	"fmt"

	"github.com/spf13/cobra"
)

var cfgFile string

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
		manager.RunForSettings(config)
		// Call method with config as an entry point
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVarP(&cfgFile, "config", "c", ".", "Path to the configuration file")
}

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"central-cyclone/cmd/dtrack"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "central-cyclone",
	Short: "A small helper for central automated SBOM creation",
	Long:  `central-cyclone allows you to create SBOMS for multiple defined repositories.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().StringP("config", "c", "./config.json", "Path to the configuration file")

	rootCmd.AddCommand(dtrack.DtCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(uploadCmd)
}

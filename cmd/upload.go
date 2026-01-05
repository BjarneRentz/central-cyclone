package cmd

import (
	"central-cyclone/internal/config"
	"central-cyclone/internal/upload"
	"central-cyclone/internal/workspace"
	"log/slog"

	"github.com/spf13/cobra"
)

var sbomFolder string

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Uploads SBOMs from a specified folder to DependencyTrack",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.LoadFromFile(cfgFile)
		if err != nil {
			slog.Error("Error loading configuration", "error", err)
			return
		}
		sbomNamer := workspace.DefaultSBOMNamer{}
		repoMapper := workspace.DefaultRepoMapper{}

		readonlyWorkspace := workspace.CreateLocalReadonlySbomWorkspace(sbomFolder, sbomNamer, repoMapper)

		sboms, err := readonlyWorkspace.ReadSboms(config.Repositories)
		if err != nil {
			slog.Error("Error reading SBOMs", "error", err)
			return
		}

		uploader, err := upload.CreateDependencyTrackUploader(config)
		if err != nil {
			slog.Error("Error creating uploader", "error", err)
			return
		}

		for _, sbom := range sboms {
			err = uploader.UploadSBOM(sbom)
			if err != nil {
				slog.Error("Error uploading SBOM", "error", err)
				continue
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringVarP(&cfgFile, "config", "c", "./config.json", "Path to the configuration file")
	uploadCmd.Flags().StringVar(&sbomFolder, "sboms-dir", "/sboms", "Directory containg the sboms to upload")
}

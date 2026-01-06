package cmd

import (
	"central-cyclone/internal/upload"
	"central-cyclone/internal/workspace"
	"log/slog"

	"github.com/spf13/cobra"
)

var sbomFolder string

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Uploads SBOMs from a specified folder to DependencyTrack",
	RunE: func(cmd *cobra.Command, args []string) error {
		settings, err := GetSettings(cmd)
		if err != nil {
			slog.Error("Could not get settings from context", "error", err)
			return err
		}

		sbomNamer := workspace.DefaultSBOMNamer{}
		repoMapper := workspace.DefaultRepoMapper{}

		readonlyWorkspace := workspace.CreateLocalReadonlySbomWorkspace(sbomFolder, sbomNamer, repoMapper)

		sboms, err := readonlyWorkspace.ReadSboms(settings.Repositories)
		if err != nil {
			slog.Error("Error reading SBOMs", "error", err)
			return err
		}

		uploader, err := upload.CreateDependencyTrackUploader(settings)
		if err != nil {
			slog.Error("Error creating uploader", "error", err)
			return err
		}

		for _, sbom := range sboms {
			err = uploader.UploadSBOM(sbom)
			if err != nil {
				slog.Error("Error uploading SBOM", "error", err)
				continue
			}
		}
		return nil
	},
}

func init() {
	requireConfig(uploadCmd)
	uploadCmd.Flags().StringP("config", "c", "./config.json", "Path to the configuration file")
	uploadCmd.Flags().StringVar(&sbomFolder, "sboms-dir", "/sboms", "Directory containg the sboms to upload")
}

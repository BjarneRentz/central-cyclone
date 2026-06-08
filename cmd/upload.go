package cmd

import (
	"context"
	"sync"

	"central-cyclone/cmd/extensions"
	"central-cyclone/internal/models"
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
		settings, err := extensions.GetSettings(cmd)
		if err != nil {
			slog.Error("Could not get settings from context", "error", err)
			return err
		}

		sbomNamer := workspace.DefaultSBOMNamer{}
		repoMapper := workspace.DefaultRepoMapper{}

		readonlyWorkspace := workspace.CreateLocalReadonlySbomWorkspace(sbomFolder, sbomNamer, repoMapper)

		sboms, err := readonlyWorkspace.ReadSboms()
		if err != nil {
			slog.Error("Error reading SBOMs", "error", err)
			return err
		}

		uploader, err := upload.CreateDependencyTrackUploader(settings)
		if err != nil {
			slog.Error("Error creating uploader", "error", err)
			return err
		}

		ctx := context.Background()

		// Limit concurrent uploads to 5
		maxConcurrency := 5
		semaphore := make(chan struct{}, maxConcurrency)

		var wg sync.WaitGroup
		var uploadErr error
		var mu sync.Mutex

		for _, sbom := range sboms {
			wg.Add(1)
			go func(s models.Sbom) {
				defer wg.Done()

				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				if err := uploader.UploadSBOM(ctx, s); err != nil {
					mu.Lock()
					slog.Error("Error uploading SBOM", "error", err)
					uploadErr = err
					mu.Unlock()
				}
			}(sbom)

		}

		wg.Wait()
		close(semaphore)

		return uploadErr
	},
}

func init() {
	extensions.RequireConfig(uploadCmd)
	uploadCmd.Flags().StringP("config", "c", "./config.json", "Path to the configuration file")
	uploadCmd.Flags().StringVar(&sbomFolder, "sboms-dir", "/sboms", "Directory containg the sboms to upload")
}

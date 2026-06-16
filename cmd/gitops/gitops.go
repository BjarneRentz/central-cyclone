package gitops

import (
	"central-cyclone/cmd/extensions"
	"central-cyclone/internal/analyzer"
	"central-cyclone/internal/config"
	"central-cyclone/internal/gitops"
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/upload"
	"central-cyclone/internal/workspace"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/spf13/cobra"
)

var GitOpsCmd = &cobra.Command{
	Use:   "gitops",
	Short: "Starts the gitops mode of central cyclone",
	RunE: func(cmd *cobra.Command, args []string) error {

		fmt.Println("⚠️ The gitops mode is currently in alpha mode")
		settings, err := extensions.GetSettings(cmd)
		if err != nil {
			slog.Error("Could not get settings from context", "error", err)
			return err
		}

		ws, err := workspace.CreateLocalWorkspace()
		if err != nil {
			slog.Error("Could not create local workspace", "error", err)
			return err
		}
		ws.Clear()

		gitTool := gittool.CreateLocalGitCloner(ws)
		configProvider := config.NewConfigProvider(settings)
		analyzer := analyzer.CdxgenAnalyzer{}
		uploader, err := upload.CreateDependencyTrackUploader(settings)
		if err != nil {
			slog.Error("Could not create DepependencyTrack Uploader", "error", err)
			return err
		}

		createSbomHandler := gitops.NewCreateSbomChangeHandler(configProvider, gitTool, analyzer, uploader)

		syncer := gitops.NewSyncer(gitTool, ws, createSbomHandler)

		err = syncer.Init(settings.GitOpsRepos)
		if err != nil {
			slog.Error("Failed to initialize syncer", "error", err)
			return err
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		refreshInterval := time.Duration(configProvider.GetGitOpsRefreshInterval()) * time.Minute
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			syncer.Reconcile()
			select {
			case <-ticker.C:
				// continue loop
			case <-sigChan:
				slog.Info("Received shutdown signal, exiting...")
				return nil
			}
		}
	},
}

func init() {
	extensions.RequireConfig(GitOpsCmd)
}

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
		ws.Clear()
		if err != nil {
			slog.Error("Could not create local workspace", "error", err)
			return err
		}

		gitTool := gittool.CreateLocalGitCloner(ws)
		configProvider := config.NewConfigProvider(settings)
		analyzer := analyzer.CdxgenAnalyzer{}
		uploader, err := upload.CreateDependencyTrackUploader(settings)

		createSbomHandler := gitops.NewCreateSbomChangeHandler(configProvider, gitTool, analyzer, uploader)

		syncer := gitops.NewSyncer(gitTool, ws, createSbomHandler)

		err = syncer.Init(settings.GitOpsRepos)
		if err != nil {
			slog.Error("Failed to initialize syncer", "error", err)
			return err
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				syncer.Reconcile()
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

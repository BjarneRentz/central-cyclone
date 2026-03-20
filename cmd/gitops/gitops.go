package gitops

import (
	"central-cyclone/cmd/extensions"
	"central-cyclone/internal/gitops"
	"central-cyclone/internal/gittool"
	"central-cyclone/internal/workspace"
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

		syncer := gitops.NewSyncer(gitTool, ws)

		err = syncer.Init(settings.GitOpsRepos)
		if err != nil {
			slog.Error("Failed to initialize syncer", "error", err)
			return err
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		ticker := time.NewTicker(5 * time.Minute)
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

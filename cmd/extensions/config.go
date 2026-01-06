package extensions

import (
	"central-cyclone/internal/config"
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

type ctxKey string

const settingsKey ctxKey = "settings"

// requireConfig sets a PreRunE on the provided command which loads the
// configuration file (using the "config" flag) and stores the parsed
// settings in the command's context.
func RequireConfig(cmd *cobra.Command) {
	prev := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if prev != nil {
			if err := prev(cmd, args); err != nil {
				return err
			}
		}

		cfg, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		settings, err := config.LoadFromFile(cfg)
		if err != nil {
			slog.Error("Could not read config file:", "error", err)
			return fmt.Errorf("could not load config: %w", err)
		}

		cmd.SetContext(context.WithValue(cmd.Context(), settingsKey, settings))
		return nil
	}
}

// GetSettings returns the loaded settings from the command's context.
// If the settings are not present it returns an error.
func GetSettings(cmd *cobra.Command) (*config.Settings, error) {
	val := cmd.Context().Value(settingsKey)
	if val == nil {
		return nil, fmt.Errorf("settings not present in context")
	}
	settings, ok := val.(*config.Settings)
	if !ok {
		return nil, fmt.Errorf("settings present in context with wrong type")
	}
	return settings, nil
}

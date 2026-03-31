package cli

import (
	"github.com/frankhildebrandt/teams2issue/internal/app"
	"github.com/frankhildebrandt/teams2issue/internal/config"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	var configPath string

	rootCmd := &cobra.Command{
		Use:   "teams2issue",
		Short: "Bootstrap daemon for teams2issue",
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", config.DefaultConfigPath, "Path to configuration file")
	rootCmd.AddCommand(newDaemonCommand(&configPath))

	return rootCmd
}

func newDaemonCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "daemon",
		Short: "Start the teams2issue daemon",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			return app.Run(cmd.Context(), cfg)
		},
	}
}

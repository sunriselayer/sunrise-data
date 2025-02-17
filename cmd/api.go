package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/sunriselayer/sunrise-data/api"
	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/context"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the API server",
	Long:  `This command starts the API server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig()
		if err != nil {
			log.Error().Msgf("Failed to load config: %s", err)
			return err
		}

		err = context.GetPublishContext(*config)
		if err != nil {
			log.Error().Msgf("Failed to connect to sunrised RPC: %s", err)
			return err
		}

		api.Handle()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)
}

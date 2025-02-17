package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/rollkit"
)

var rollkitCmd = &cobra.Command{
	Use:   "rollkit",
	Short: "Start the Rollkit server",
	Long:  `This command starts the Rollkit server.`,
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

		rollkit.Serve()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rollkitCmd)
}

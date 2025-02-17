package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/context"
)

var rollkitCmd = &cobra.Command{
	Use:   "rollkit",
	Short: "Start the Rollkit server",
	Long:  `This command starts the Rollkit server.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.LoadConfig()
		if err != nil {
			log.Error().Msgf("Failed to load config: %s", err)
			return
		}

		err = context.GetPublishContext(*config)
		if err != nil {
			log.Error().Msgf("Failed to connect to sunrised RPC: %s", err)
			return
		}

		Serve()
	},
}

func init() {
	rootCmd.AddCommand(rollkitCmd)
}

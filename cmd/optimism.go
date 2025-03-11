package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/sunriselayer/sunrise-data/config"
	appctx "github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/optimism"
	"github.com/sunriselayer/sunrise-data/protocols"
)

var optimismCmd = &cobra.Command{
	Use:   "optimism",
	Short: "Start Optimism DA Server",
	Long:  `This command starts the Optimism DA Server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig()
		if err != nil {
			log.Error().Msgf("Failed to load config: %s", err)
			return err
		}

		if err = appctx.GetPublishContext(*config); err != nil {
			log.Error().Msgf("Failed to connect to sunrised RPC: %s", err)
			return err
		}

		if err := protocols.CheckIpfsConnection(); err != nil {
			log.Error().Msgf("Failed to connect to IPFS: %s", err)
			return err
		}

		if err := optimism.StartDAServer(); err != nil {
			log.Error().Msgf("Failed to start Optimism DA Server: %s", err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(optimismCmd)
}

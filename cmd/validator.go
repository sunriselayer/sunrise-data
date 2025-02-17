package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/tasks"
)

var validatorCmd = &cobra.Command{
	Use:   "validator",
	Short: "Run validator's proof tasks",
	Long:  `This command starts the validator's proof tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.LoadConfig()
		if err != nil {
			log.Error().Msgf("Failed to load config: %s", err)
			return
		}

		// TODO check rpc is enabled
		err = context.GetProofContext(*config)
		if err != nil {
			log.Error().Msgf("Failed to connect to sunrised RPC: %s", err)
			return
		}

		tasks.RunTasks()
	},
}

func init() {
	rootCmd.AddCommand(validatorCmd)
}

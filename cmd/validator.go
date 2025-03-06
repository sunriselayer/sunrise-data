package cmd

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/protocols"
	"github.com/sunriselayer/sunrise-data/validator"
)

var validatorCmd = &cobra.Command{
	Use:   "validator",
	Short: "Run validator's proof tasks",
	Long:  `This command starts the validator's proof tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.LoadConfig()
		if err != nil {
			log.Error().Msgf("Failed to load config: %s", err)
			return err
		}

		if err = context.GetProofContext(*config); err != nil {
			log.Error().Msgf("Failed to connect to sunrised RPC: %s", err)
			return err
		}

		if err := protocols.CheckIpfsConnection(); err != nil {
			log.Error().Msgf("Failed to connect to IPFS: %s", err)
			return err
		}

		ok := validator.RunValidatorTask()
		if !ok {
			return errors.New("failed to start validator task")
		}

		// Block forever to keep the validator running
		<-make(chan struct{})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validatorCmd)
}

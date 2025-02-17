package validator

import (
	"github.com/rs/zerolog/log"
	"github.com/sunriselayer/sunrise-data/context"
	datypes "github.com/sunriselayer/sunrise/x/da/types"
)

// RunValidatorTask is a function to run threads.
func RunValidatorTask() bool {
	log.Info().Msg("Starting validator task")
	validatorAddress := context.Config.Validator.ValidatorAddress
	deputyAddress := context.Addr
	log.Info().Msgf("validator: %s deputy: %s", validatorAddress, deputyAddress)
	res, err := context.QueryClient.ProofDeputy(context.Ctx, &datypes.QueryProofDeputyRequest{ValidatorAddress: validatorAddress})
	if err != nil {
		log.Error().Msgf("Failed to query proof deputy: %s", err)
		log.Info().Msg("Please send a MsgRegisterProofDeputy tx from your validator address")
		return false
	}
	if res.DeputyAddress != deputyAddress {
		log.Error().Msgf("%s is not registered as a proof deputy of %s", deputyAddress, validatorAddress)
		log.Info().Msg("Please send a MsgRegisterProofDeputy tx from your validator address")
		return false
	}
	go Monitor()
	return true
}

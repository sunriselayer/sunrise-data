package optimism

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-service/opio"
	"github.com/rs/zerolog/log"
	"github.com/sunriselayer/sunrise-data/config"
)

type Server interface {
	Start() error
	Stop() error
}

func StartDAServer() error {
	log.Info().Msg("Initializing Alt DA Server...")
	config, err := config.LoadConfig()
	if err != nil {
		return err
	}

	log.Info().Msgf("Starting Alt DA Server... listen_address: %s, port: %d, data_shard_count: %d, parity_shard_count: %d", config.Optimism.ListenAddress, config.Optimism.Port, config.Optimism.DataShardCount, config.Optimism.ParityShardCount)
	storeConfig := SunriseConfig{
		DataShardCount:   config.Optimism.DataShardCount,
		ParityShardCount: config.Optimism.ParityShardCount,
	}
	store := NewSunriseStore(storeConfig)
	server := NewSunriseServer(config.Optimism.ListenAddress, config.Optimism.Port, store)

	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start the Alt DA Server")
	} else {
		log.Info().Msg("Started Alt DA Server")
	}

	defer func() {
		if err := server.Stop(); err != nil {
			log.Error().Msgf("failed to stop Alt DA Server: %s", err)
		}
	}()

	opio.BlockOnInterrupts()

	return nil
}

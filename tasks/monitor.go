package tasks

import (
	"encoding/base64"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
	datypes "github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/protocols"
	"github.com/sunriselayer/sunrise-data/utils"
)

func Monitor() {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				MonitorChallengingData()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func MonitorChallengingData() {
	if context.Config.Chain.ValidatorAddress == "" {
		log.Error().Msg("validator_address is empty in config.toml")
		return
	}
	allPublishedDataResponse, err := context.QueryClient.AllPublishedData(context.Ctx, &datypes.QueryAllPublishedDataRequest{})
	if err != nil {
		log.Error().Msgf("Failed to query all-published-data from on-chain: %s", err)
		return
	}
	allData := allPublishedDataResponse.Data
	for _, data := range allData {
		if data.Status == datypes.Status_STATUS_CHALLENGING {
			_, err := context.QueryClient.ValidityProof(context.Ctx, &datypes.QueryValidityProofRequest{MetadataUri: data.MetadataUri, ValidatorAddress: context.Config.Chain.ValidatorAddress})
			if err != nil {
				SubmitProofTx(data)
			}
			continue
		}
	}
}

func SubmitProofTx(data datypes.PublishedData) bool {
	peerAddrInfo, err := peer.AddrInfoFromString(data.DataSourceInfo)
	if err == nil {
		protocols.ConnectSwarm(*peerAddrInfo)
	}
	protocol, err := protocols.GetRetrieveProtocol(data.MetadataUri)
	if err != nil {
		log.Error().Msgf("Failed to get protocol: %s", err)
		return false
	}

	// verify shard data
	metadataBytes, err := protocol.Retrieve(data.MetadataUri)
	if err != nil {
		log.Error().Msgf("Failed to get metadata: %s", err)
		return false
	}
	metadata := datypes.Metadata{}
	if err := metadata.Unmarshal(metadataBytes); err != nil {
		log.Error().Msgf("Failed to decode metadata: %s", err)
		return false
	}

	if len(data.ShardDoubleHashes) != len(metadata.ShardUris) {
		log.Error().Msgf("Incorrect shard data count: %d %d", len(data.ShardDoubleHashes), len(metadata.ShardUris))
		return false
	}

	validShards := [][]byte{}
	validShardIndexes := []int{}

	for index, doubleHash := range data.ShardDoubleHashes {
		shardUri := metadata.ShardUris[index]
		shardData, err := protocol.Retrieve(shardUri)
		if err != nil {
			log.Error().Msgf("Failed to get shard data: %s", err)
			continue
		}

		doubleShardHash := base64.StdEncoding.EncodeToString(utils.DoubleHashMimc(shardData))
		if doubleShardHash != base64.StdEncoding.EncodeToString(doubleHash) {
			log.Error().Msgf("Incorrect shard data: %d", index)
			continue
		}
		validShards = append(validShards, shardData)
		validShardIndexes = append(validShardIndexes, index)
	}

	DataShardCount := len(data.ShardDoubleHashes) - int(metadata.ParityShardCount)

	if len(validShards) < DataShardCount {
		log.Error().Msgf("Valid shard count less than DataShardCount: %d", len(validShards))
		return false
	}

	shardLength := len(metadata.ShardUris)
	queryThresholdResponse, err := context.QueryClient.ZkpProofThreshold(context.Ctx, &datypes.QueryZkpProofThresholdRequest{ShardCount: uint64(shardLength)})
	if err != nil {
		log.Error().Msgf("Failed to query Threshold: %s", err)
		return false
	}

	threshold := queryThresholdResponse.Threshold
	validator, err := sdk.ValAddressFromBech32(context.Config.Chain.ValidatorAddress)
	if err != nil {
		log.Error().Msgf("Failed to parse ValidatorAddress: %s %s", context.Config.Chain.ValidatorAddress, err)
		return false
	}

	requiredIndices := datypes.ShardIndicesForValidator(validator, int64(threshold), int64(shardLength))
	proofs := [][]byte{}
	indices := []int64{}

	for _, index := range requiredIndices {
		for i, validIndex := range validShardIndexes {
			if index != int64(validIndex) {
				continue
			}

			shardData := validShards[i]
			shardHash := utils.HashMimc(shardData)
			doubleShardHash := utils.HashMimc(shardHash)
			proofBytes, ok := getShardProofBytes(shardHash, doubleShardHash)
			if !ok {
				log.Error().Msgf("Failed to generate shard proof: %s, indice: %d", data.MetadataUri, index)
				return false
			}

			proofs = append(proofs, proofBytes)
			indices = append(indices, index)
		}
	}

	return submitValidityProof(data.MetadataUri, indices, proofs)
}

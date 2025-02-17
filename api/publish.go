package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/sunriselayer/sunrise/x/da/erasurecoding"
	"github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/protocols"
	"github.com/sunriselayer/sunrise-data/utils"
)

func Publish(w http.ResponseWriter, r *http.Request) {
	var req PublishRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := PublishData(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Info().Msgf("TxHash: %s", res.TxHash)
	// Print response from broadcasting a transaction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func PublishData(req PublishRequest) (PublishResponse, error) {
	blobBytes, err := base64.StdEncoding.DecodeString(req.Blob)
	if err != nil {
		log.Err(err).Msg("Failed to decode blob")
		return PublishResponse{}, err
	}

	publishProtocol, err := protocols.GetPublishProtocol(req.Protocol)
	if err != nil {
		log.Err(err).Msg("Failed to get publish protocol")
		return PublishResponse{}, err
	}

	recoveredDataHash, err := utils.HashSha256(blobBytes)
	if err != nil {
		log.Err(err).Msg("Failed to hash blob")
		return PublishResponse{}, err
	}

	queryClient := types.NewQueryClient(context.NodeClient.Context())
	queryParamResponse, err := queryClient.Params(context.Ctx, &types.QueryParamsRequest{})
	if err != nil {
		log.Err(err).Msg("Failed to query da params")
		return PublishResponse{}, err
	}
	if queryParamResponse.Params.MinShardCount > uint64(req.DataShardCount+req.ParityShardCount) {
		log.Error().Msg("DataShardCount + ParityShardCount is smaller than Min_ShardCount")
		return PublishResponse{}, errors.New("DataShardCount + ParityShardCount is smaller than Min_ShardCount")
	}
	if queryParamResponse.Params.MaxShardCount < uint64(req.DataShardCount+req.ParityShardCount) {
		log.Error().Msg("DataShardCount + ParityShardCount is bigger than Max_ShardCount")
		return PublishResponse{}, errors.New("DataShardCount + ParityShardCount is bigger than Max_ShardCount")
	}

	shardSize, _, shards, err := erasurecoding.ErasureCode(blobBytes, req.DataShardCount, req.ParityShardCount)
	if err != nil {
		log.Err(err).Msg("Failed to erasure code")
		return PublishResponse{}, err
	}
	if queryParamResponse.Params.MaxShardSize < shardSize {
		log.Error().Msg("ShardSize is bigger than Max_ShardSize")
		return PublishResponse{}, errors.New("ShardSize is bigger than Max_ShardSize")
	}
	shardUris, err := publishProtocol.PublishShards(shards)
	if err != nil {
		log.Err(err).Msg("Failed to publish shards")
		return PublishResponse{}, err
	}
	metadata := types.Metadata{
		ShardSize:         shardSize,
		RecoveredDataHash: recoveredDataHash,
		RecoveredDataSize: uint64(len(blobBytes)),
		ShardUris:         shardUris,
	}
	metadataBytes, err := metadata.Marshal()
	if err != nil {
		log.Err(err).Msg("Failed to marshal metadata")
		return PublishResponse{}, err
	}

	metadataUri, err := publishProtocol.PublishMetadata(metadataBytes)
	if err != nil {
		log.Err(err).Msg("Failed to publish metadata")
		return PublishResponse{}, err
	}

	// Define a message to create a post
	msg := &types.MsgPublishData{
		Sender:            context.Addr,
		MetadataUri:       metadataUri,
		ParityShardCount:  uint64(req.ParityShardCount),
		ShardDoubleHashes: utils.ByteSlicesToDoubleHashes(shards),
		DataSourceInfo:    context.Config.Api.IpfsAddressInfo,
	}
	// Broadcast a transaction from account `alice` with the message
	// to create a post store response in txResp
	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, msg)
	if err != nil {
		log.Err(err).Msg("Failed to broadcast tx")
		return PublishResponse{}, err
	}
	log.Info().Msgf("TxHash: %s", txResp.TxHash)
	return PublishResponse{
		TxHash:      txResp.TxHash,
		MetadataUri: metadataUri,
	}, nil
}

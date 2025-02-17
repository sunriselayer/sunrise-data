package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/sunriselayer/sunrise/x/da/erasurecoding"
	"github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/protocols"
	"github.com/sunriselayer/sunrise-data/utils"
)

func PublishFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	fileName := r.FormValue("file_name")
	protocol := r.FormValue("protocol")

	publishProtocol, err := protocols.GetPublishProtocol(protocol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dataShardCount, err := strconv.Atoi(r.FormValue("data_shard_count"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parityShardCount, err := strconv.Atoi(r.FormValue("parity_shard_count"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile(fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Error().Msgf("Failed to read file %s", fileName)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recoveredDataHash, err := utils.HashSha256(fileBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queryClient := types.NewQueryClient(context.NodeClient.Context())
	queryParamResponse, err := queryClient.Params(context.Ctx, &types.QueryParamsRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if queryParamResponse.Params.MinShardCount > uint64(dataShardCount+parityShardCount) {
		http.Error(w, "DataShardCount + ParityShardCount is less than Min_ShardCount", http.StatusBadRequest)
		return
	}
	if queryParamResponse.Params.MaxShardCount < uint64(dataShardCount+parityShardCount) {
		http.Error(w, "DataShardCount + ParityShardCount is bigger than Max_ShardCount", http.StatusBadRequest)
		return
	}

	shardSize, _, shards, err := erasurecoding.ErasureCode(fileBytes, dataShardCount, parityShardCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if queryParamResponse.Params.MaxShardSize < shardSize {
		http.Error(w, "ShardSize is bigger than Max_ShardSize", http.StatusBadRequest)
		return
	}
	shardUris, err := publishProtocol.PublishShards(shards)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	metadata := types.Metadata{
		ShardSize:         shardSize,
		RecoveredDataHash: recoveredDataHash,
		RecoveredDataSize: uint64(len(fileBytes)),
		ShardUris:         shardUris,
	}
	metadataBytes, err := metadata.Marshal()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadataUri := ""
	metadataUri, err = publishProtocol.PublishMetadata(metadataBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Define a message to create a post
	msg := &types.MsgPublishData{
		Sender:            context.Addr,
		MetadataUri:       metadataUri,
		ParityShardCount:  uint64(parityShardCount),
		ShardDoubleHashes: utils.ByteSlicesToDoubleHashes(shards),
		DataSourceInfo:    context.Config.Api.IpfsAddressInfo,
	}
	// Broadcast a transaction from account `alice` with the message
	// to create a post store response in txResp
	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info().Msgf("TxHash: %s", txResp.TxHash)
	// Print response from broadcasting a transaction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PublishResponse{
		TxHash:      txResp.TxHash,
		MetadataUri: metadataUri,
	})
}

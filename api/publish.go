package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/sunriselayer/sunrise/x/da/erasurecoding"
	"github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/consts"
	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/publisher"
	"github.com/sunriselayer/sunrise-data/utils"
)

func Publish(w http.ResponseWriter, r *http.Request) {
	var req PublishRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	blobBytes, err := base64.StdEncoding.DecodeString(req.Blob)
	if err != nil {
		log.Error().Msgf("Failed to decode Blob %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	protocol := consts.IPFS_PROTOCOL
	if req.Protocol == consts.IPFS_PROTOCOL {
		protocol = consts.IPFS_PROTOCOL
	} else if req.Protocol == consts.ARWEAVE_PROTOCOL {
		protocol = consts.ARWEAVE_PROTOCOL
	} else {
		http.Error(w, "Invalid Protocol", http.StatusBadRequest)
		return
	}

	recoveredDataHash, err := utils.HashSha256(blobBytes)
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
	if queryParamResponse.Params.MinShardCount > uint64(req.DataShardCount+req.ParityShardCount) {
		http.Error(w, "DataShardCount + ParityShardCount is less than Min_ShardCount", http.StatusBadRequest)
		return
	}
	if queryParamResponse.Params.MaxShardCount < uint64(req.DataShardCount+req.ParityShardCount) {
		http.Error(w, "DataShardCount + ParityShardCount is bigger than Max_ShardCount", http.StatusBadRequest)
		return
	}

	shardSize, _, shards, err := erasurecoding.ErasureCode(blobBytes, req.DataShardCount, req.ParityShardCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if queryParamResponse.Params.MaxShardSize < shardSize {
		http.Error(w, "ShardSize is bigger than Max_ShardSize", http.StatusBadRequest)
		return
	}
	shardUris, err := publisher.GetShardUris(shards, protocol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	metadata := types.Metadata{
		ShardSize:         shardSize,
		RecoveredDataHash: recoveredDataHash,
		RecoveredDataSize: uint64(len(blobBytes)),
		ShardUris:         shardUris,
	}
	metadataBytes, err := metadata.Marshal()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//upload ipfs
	metadataUri := ""
	if protocol == consts.IPFS_PROTOCOL {
		metadataUri, err = publisher.UploadToIpfs(metadataBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		metadataUri, err = publisher.UploadToArweave(metadataBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Define a message to create a post
	msg := &types.MsgPublishData{
		Sender:            context.Addr,
		MetadataUri:       metadataUri,
		ParityShardCount:  uint64(req.ParityShardCount),
		ShardDoubleHashes: utils.ByteSlicesToDoubleHashes(shards),
		DataSourceInfo:    context.Config.Api.IpfsAddrInfo,
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

package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sunriselayer/sunrise-data/cosmosclient"
	"github.com/sunriselayer/sunrise/x/da/erasurecoding"
	"github.com/sunriselayer/sunrise/x/da/types"
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	protocol := IPFS_PROTOCOL
	if req.Protocol == IPFS_PROTOCOL {
		protocol = IPFS_PROTOCOL
	} else if req.Protocol == ARWEAVE_PROTOCOL {
		protocol = ARWEAVE_PROTOCOL
	} else {
		http.Error(w, "Invalid Protocol", http.StatusBadRequest)
		return
	}

	recoveredDataHash, err := HashSha256(blobBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	nodeClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(SUNRISE_ADDR_PRIFIX))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	queryClient := types.NewQueryClient(nodeClient.Context())
	queryParamResponse, err := queryClient.Params(ctx, &types.QueryParamsRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if queryParamResponse.Params.MinShardCount > uint64(req.ShardCountHalf*2) {
		http.Error(w, "ShardCount is less than Min_ShardCount", http.StatusBadRequest)
		return
	}

	shardSize, _, shards := erasurecoding.ErasureCode(blobBytes, req.ShardCountHalf)
	if queryParamResponse.Params.MaxShardSize < shardSize {
		http.Error(w, "ShardSize is bigger than Max_ShardSize", http.StatusBadRequest)
		return
	}
	shardUris, err := GetShardUris(shards, protocol)
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
	if protocol == IPFS_PROTOCOL {
		metadataUri, err = UploadToIpfs(metadataBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		metadataUri, err = UploadToArweave(metadataBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Account `alice` was initialized during `ignite chain serve`
	accountName := "validator"
	// Get account from the keyring
	account, err := nodeClient.Account(accountName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addr, err := account.Address(SUNRISE_ADDR_PRIFIX)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Define a message to create a post
	msg := &types.MsgPublishData{
		Sender:            addr,
		MetadataUri:       metadataUri,
		ShardDoubleHashes: byteSlicesToDoubleHashes(shards),
	}
	// Broadcast a transaction from account `alice` with the message
	// to create a post store response in txResp
	txResp, err := nodeClient.BroadcastTx(ctx, account, msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("TxHash:", txResp.TxHash)
	// Print response from broadcasting a transaction
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PublishResponse{
		MetadataUri: metadataUri,
	})
}

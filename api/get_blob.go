package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/sunriselayer/sunrise/x/da/erasurecoding"
	"github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/protocols"
)

func GetBlob(w http.ResponseWriter, r *http.Request) {
	metadataUri := r.URL.Query().Get("metadata_uri")
	if metadataUri == "" {
		http.Error(w, "Invalid query parameter", http.StatusBadRequest)
		return
	}

	res, err := GetBlobData(metadataUri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func GetBlobData(metadataUri string) (GetBlobResponse, error) {
	protocol, err := protocols.GetRetrieveProtocol(metadataUri)
	if err != nil {
		return GetBlobResponse{}, err
	}

	metadataBytes, err := protocol.Retrieve(metadataUri)
	if err != nil {
		return GetBlobResponse{}, err
	}

	metadata := types.Metadata{}

	if err := metadata.Unmarshal(metadataBytes); err != nil {
		return GetBlobResponse{}, err
	}
	shardUris := metadata.ShardUris
	var shards [][]byte

	for _, shardUri := range shardUris {
		shardData, err := protocol.Retrieve(shardUri)
		if err != nil {
			log.Error().Msgf("Failed to retrieve shard uri %s", err)
			continue
		}
		shards = append(shards, shardData)
	}
	DataShardCount := len(shardUris) - int(metadata.ParityShardCount)
	blobBytes, err := erasurecoding.JoinShards(shards, DataShardCount, int(metadata.RecoveredDataSize))
	if err != nil {
		return GetBlobResponse{}, err
	}
	res := GetBlobResponse{
		Blob: base64.StdEncoding.EncodeToString(blobBytes),
	}
	return res, nil
}

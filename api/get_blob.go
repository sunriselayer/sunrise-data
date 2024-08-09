package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/sunriselayer/sunrise-data/retriever"
	"github.com/sunriselayer/sunrise/x/da/erasurecoding"
	"github.com/sunriselayer/sunrise/x/da/types"
)

func GetBlob(w http.ResponseWriter, r *http.Request) {
	metadataUri := r.URL.Query().Get("metadata_uri")
	if metadataUri == "" {
		http.Error(w, "Invalid query parameter", http.StatusBadRequest)
		return
	}
	metadataBytes, err := retriever.GetDataFromIpfsOrArweave(metadataUri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata := types.Metadata{}

	if err := metadata.Unmarshal(metadataBytes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	shardUris := metadata.ShardUris
	var shards [][]byte

	for _, shardUri := range shardUris {
		shardData, err := retriever.GetDataFromIpfsOrArweave(shardUri)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		shards = append(shards, shardData)
	}
	blobBytes, err := erasurecoding.JoinShards(shards, int(metadata.RecoveredDataSize))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res := GetBlobResponse{
		Blob: base64.StdEncoding.EncodeToString(blobBytes),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

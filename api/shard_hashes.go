package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/retriever"
	"github.com/sunriselayer/sunrise-data/utils"
)

func ShardHashes(w http.ResponseWriter, r *http.Request) {
	metadataUri := r.URL.Query().Get("metadata_uri")
	indices := r.URL.Query().Get("indices")
	indicesList := strings.Split(indices, ",")
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
	shardHashes := []string{}

	for _, index := range indicesList {
		iIndex, err := strconv.Atoi(index)
		if err != nil {
			continue
		}
		if iIndex < len(metadata.ShardUris) {
			shardUri := metadata.ShardUris[iIndex]
			shardData, err := retriever.GetDataFromIpfsOrArweave(shardUri)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			shardHashes = append(shardHashes, base64.StdEncoding.EncodeToString(utils.HashMimc(shardData)))
		}
	}

	res := UploadedDataResponse{
		ShardSize:   metadata.ShardSize,
		ShardUris:   metadata.ShardUris,
		ShardHashes: shardHashes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

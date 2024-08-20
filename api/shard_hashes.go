package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/retriever"
	"github.com/sunriselayer/sunrise-data/utils"
)

func ShardHashes(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ShardHashes-1", time.Now())
	metadataUri := r.URL.Query().Get("metadata_uri")
	indices := r.URL.Query().Get("indices")
	indicesList := strings.Split(indices, ",")
	if metadataUri == "" {
		http.Error(w, "Invalid query parameter", http.StatusBadRequest)
		return
	}
	fmt.Println("ShardHashes-2", time.Now())
	metadataBytes, err := retriever.GetDataFromIpfsOrArweave(metadataUri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("ShardHashes-3", time.Now())

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
			fmt.Println("ShardHashes-4", time.Now())
			shardUri := metadata.ShardUris[iIndex]
			shardData, err := retriever.GetDataFromIpfsOrArweave(shardUri)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			fmt.Println("ShardHashes-5", time.Now())
			shardHashes = append(shardHashes, base64.StdEncoding.EncodeToString(utils.HashMimc(shardData)))
		}
	}

	res := UploadedDataResponse{
		ShardSize:   metadata.ShardSize,
		ShardUris:   metadata.ShardUris,
		ShardHashes: shardHashes,
	}

	fmt.Println("ShardHashes-6", time.Now())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

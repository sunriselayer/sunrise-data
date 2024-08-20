package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/retriever"
	"github.com/sunriselayer/sunrise-data/utils"
)

func ShardHashes(w http.ResponseWriter, r *http.Request) {
	log.Info().Msgf("ShardHashes-1 %d", time.Now())
	metadataUri := r.URL.Query().Get("metadata_uri")
	indices := r.URL.Query().Get("indices")
	indicesList := strings.Split(indices, ",")
	if metadataUri == "" {
		http.Error(w, "Invalid query parameter", http.StatusBadRequest)
		return
	}
	log.Info().Msgf("ShardHashes-2 %d", time.Now())
	metadataBytes, err := retriever.GetDataFromIpfsOrArweave(metadataUri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Info().Msgf("ShardHashes-3 %d", time.Now())

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
			log.Info().Msgf("ShardHashes-4 %d", time.Now())
			shardUri := metadata.ShardUris[iIndex]
			shardData, err := retriever.GetDataFromIpfsOrArweave(shardUri)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			log.Info().Msgf("ShardHashes-5 %d", time.Now())
			shardHashes = append(shardHashes, base64.StdEncoding.EncodeToString(utils.HashMimc(shardData)))
		}
	}

	res := UploadedDataResponse{
		ShardSize:   metadata.ShardSize,
		ShardUris:   metadata.ShardUris,
		ShardHashes: shardHashes,
	}

	log.Info().Msgf("ShardHashes-6 %d", time.Now())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

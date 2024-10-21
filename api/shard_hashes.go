package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/protocols"
	"github.com/sunriselayer/sunrise-data/utils"
)

type ShardDataReponse struct {
	shard []byte
	index int
}

func ShardHashes(w http.ResponseWriter, r *http.Request) {
	metadataUri := r.URL.Query().Get("metadata_uri")
	indices := r.URL.Query().Get("indices")
	indicesList := strings.Split(indices, ",")
	if metadataUri == "" {
		http.Error(w, "Invalid query parameter", http.StatusBadRequest)
		return
	}

	protocol, err := protocols.GetRetrieveProtocol(metadataUri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadataBytes, err := protocol.Retrieve(metadataUri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata := types.Metadata{}
	if err := metadata.Unmarshal(metadataBytes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shardDataResponseCh := make(chan ShardDataReponse, len(indicesList))
	for i, indiceIndex := range indicesList {
		go func() {
			iIndex, err := strconv.Atoi(indiceIndex)
			if err != nil {
				return
			}
			if iIndex < len(metadata.ShardUris) {
				shardUri := metadata.ShardUris[iIndex]
				shardData, err := protocol.Retrieve(shardUri)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				shardDataResponseCh <- ShardDataReponse{shardData, i}
			}
		}()
	}

	var shardDataResponses []ShardDataReponse
	for range indicesList {
		chValue := <-shardDataResponseCh
		shardDataResponses = append(shardDataResponses, chValue)
	}
	sort.Slice(shardDataResponses, func(i, j int) bool {
		return shardDataResponses[i].index < shardDataResponses[j].index
	})

	shardHashes := []string{}
	for _, shardData := range shardDataResponses {
		shardHashes = append(shardHashes, base64.StdEncoding.EncodeToString(utils.HashMimc(shardData.shard)))
	}

	res := UploadedDataResponse{
		ShardSize:   metadata.ShardSize,
		ShardUris:   metadata.ShardUris,
		ShardHashes: shardHashes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

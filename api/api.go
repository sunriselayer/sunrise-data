package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	scontext "github.com/sunriselayer/sunrise-data/context"
)

type UploadedDataResponse struct {
	ShardSize   uint64   `json:"shard_size"`
	ShardUris   []string `json:"shard_uris"`
	ShardHashes []string `json:"shard_hashes"`
}

type PublishRequest struct {
	Blob             string `json:"blob"`
	DataShardCount   int    `json:"data_shard_count"`
	ParityShardCount int    `json:"parity_shard_count"`
	Protocol         string `json:"protocol"`
}

type PublishResponse struct {
	TxHash      string `json:"tx_hash"`
	MetadataUri string `json:"metadata_uri"`
}

type GetBlobResponse struct {
	Blob string `json:"blob"`
}

func Handle() {
	r := mux.NewRouter()
	r.HandleFunc("/publish", Publish).Methods("POST")

	r.HandleFunc("/shard-hashes", ShardHashes).Methods("GET")
	r.HandleFunc("/blob", GetBlob).Methods("GET")

	log.Info().Msgf("Running Publisher API on localhost: %d", scontext.Config.Api.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", scontext.Config.Api.Port), r)

}

package api

import (
	"bytes"
	"context"
	"encoding/json"

	"net/http"

	"encoding/base64"

	"crypto/sha256"

	"github.com/gorilla/mux"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/sunriselayer/sunrise-data/cosmosclient"
	"github.com/sunriselayer/sunrise/x/da/erasurecoding"
	"github.com/sunriselayer/sunrise/x/da/types"
)

type UploadedDataResponse struct {
	ShardHashes []string `json:"shard_hashes"`
}

type PublishRequest struct {
	Blob string `json:"blob"`
}

type PublishResponse struct {
	TxHash string `json:"tx_hash"`
}

const SUNRISE_ADDR_PRIFIX string = "sunrise"

func Handle() {
	r := mux.NewRouter()
	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to cau-sunrise-data API"))
	})
	r.HandleFunc("/api/publish", func(w http.ResponseWriter, r *http.Request) {
		var req PublishRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		base64Bytes, err := base64.StdEncoding.DecodeString(req.Blob)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		shardCountHalf := 2 //how to set?
		//how to use shard_size, shard_count
		shardSize, shardCount, shards := erasurecoding.ErasureCode(base64Bytes, shardCountHalf)
		metadata := types.Metadata{
			ShardSize:  shardSize,
			ShardCount: uint64(shardCount),
			ShardUris:  byteSlicesToBase64Strings(shards),
		}
		metadataBytes, err := metadata.Marshal()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		//upload ipfs
		node, err := rpc.NewLocalApi()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx := context.Background()
		// Create a reader for the byte array
		reader := bytes.NewReader(metadataBytes)
		fileReader := files.NewReaderFile(reader)
		cidFile, err := node.Unixfs().Add(ctx, fileReader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		nodeClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(SUNRISE_ADDR_PRIFIX))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
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
			RecoveredDataHash: []byte{},
			MetadataUri:       cidFile.String(),
			ShardDoubleHashes: byteSlicesToDoubleHashes(shards),
		}
		// Broadcast a transaction from account `alice` with the message
		// to create a post store response in txResp
		txResp, err := nodeClient.BroadcastTx(ctx, account, msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Print response from broadcasting a transaction
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PublishResponse{
			TxHash: txResp.TxHash,
		})
	}).Methods("POST")

	r.HandleFunc("/api/uploaded_data", func(w http.ResponseWriter, r *http.Request) {
		metadataUri := r.URL.Query().Get("metadata_uri")
		if metadataUri == "" {
			http.Error(w, "Invalid query parameter", http.StatusBadRequest)
			return
		}

		ctx := context.Background()

		// Create a Cosmos client instance
		nodeClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(SUNRISE_ADDR_PRIFIX))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		queryClient := types.NewQueryClient(nodeClient.Context())
		queryResp, err := queryClient.PublishedData(ctx, &types.QueryPublishedDataRequest{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		publishedDataList := queryResp.GetData()
		res := UploadedDataResponse{}

		for _, publishedData := range publishedDataList {
			if publishedData.MetadataUri == metadataUri {
				res.ShardHashes = byteSlicesToBase64Strings(publishedData.ShardDoubleHashes)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)

	}).Methods("GET")
	http.ListenAndServe(":8080", r)
}

func byteSlicesToBase64Strings(inputData [][]byte) []string {
	var base64Strings []string
	for _, data := range inputData {
		base64String := base64.StdEncoding.EncodeToString(data)
		base64Strings = append(base64Strings, base64String)
	}
	return base64Strings
}

func byteSlicesToDoubleHashes(inputData [][]byte) [][]byte {
	var convertedData [][]byte
	for _, data := range inputData {
		convertedData = append(convertedData, doubleHashSha256(data))
	}
	return convertedData
}

func doubleHashSha256(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	bs := h.Sum(nil)
	h.Reset()
	h.Write(bs)
	return h.Sum(nil)
}

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"net/http"

	"encoding/base64"

	native_mimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/gorilla/mux"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/sunriselayer/sunrise-data/cosmosclient"
	"github.com/sunriselayer/sunrise/x/da/erasurecoding"
	"github.com/sunriselayer/sunrise/x/da/types"
)

type UploadedDataResponse struct {
	ShardHashes []string `json:"shard_hashes"`
}

type PublishRequest struct {
	Blob           string `json:"blob"`
	ShardCountHalf int    `json:"shard_count_half"`
}

type PublishResponse struct {
	MetadataUri string `json:"metadata_uri"`
}

const SUNRISE_ADDR_PRIFIX string = "sunrise"

func Handle() {
	r := mux.NewRouter()
	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to cau-sunrise-data API"))
	}).Methods("GET")
	r.HandleFunc("/api/publish", func(w http.ResponseWriter, r *http.Request) {
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

		shardSize, shardCount, shards := erasurecoding.ErasureCode(blobBytes, req.ShardCountHalf)
		shardUris, err := getShardUris(shards)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		metadata := types.Metadata{
			ShardSize:  shardSize,
			ShardCount: uint64(shardCount),
			ShardUris:  shardUris,
		}
		metadataBytes, err := metadata.Marshal()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		//upload ipfs
		metadataUri, err := uploadIpfs(metadataBytes)
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
			RecoveredDataHash: hashMimc(blobBytes),
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
	}).Methods("POST")

	r.HandleFunc("/api/uploaded_data", func(w http.ResponseWriter, r *http.Request) {
		metadataUri := r.URL.Query().Get("metadata_uri")
		if metadataUri == "" {
			http.Error(w, "Invalid query parameter", http.StatusBadRequest)
			return
		}
		metadataBytes, err := getDataFromIpfs(metadataUri)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metadata := types.Metadata{}

		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		shardHashes := []string{}
		for _, shardUri := range metadata.ShardUris {
			shard, err := getDataFromIpfs(shardUri)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			shardHashes = append(shardHashes, base64.StdEncoding.EncodeToString(hashMimc(shard)))
		}

		res := UploadedDataResponse{
			ShardHashes: shardHashes,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)

	}).Methods("GET")
	fmt.Println("Running on localhost:8080")
	http.ListenAndServe(":8080", r)

}

func byteSlicesToDoubleHashes(inputData [][]byte) [][]byte {
	var convertedData [][]byte
	for _, data := range inputData {
		convertedData = append(convertedData, doubleHashMimc(data))
	}
	return convertedData
}

func hashMimc(data []byte) []byte {
	m := native_mimc.NewMiMC()
	m.Write(data)
	return m.Sum(nil)
}

func doubleHashMimc(data []byte) []byte {
	hashData := hashMimc(data)
	return hashMimc(hashData)
}

// upload shards to ipfs
func getShardUris(inputData [][]byte) ([]string, error) {
	var shardUris []string
	for _, data := range inputData {
		shardUri, err := uploadIpfs(data)
		if err != nil {
			return nil, err
		}
		shardUris = append(shardUris, shardUri)
	}
	return shardUris, nil
}

func uploadIpfs(inputData []byte) (string, error) {
	node, err := rpc.NewLocalApi()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	reader := bytes.NewReader(inputData)
	fileReader := files.NewReaderFile(reader)
	cidFile, err := node.Unixfs().Add(ctx, fileReader)
	if err != nil {
		return "", err
	}
	return cidFile.String(), nil
}

func getDataFromIpfs(uri string) ([]byte, error) {
	node, err := rpc.NewLocalApi()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	cidData, err := cid.Parse(uri)
	if err != nil {
		return nil, err
	}
	data, err := node.Unixfs().Get(ctx, path.FromCid(cidData))
	if err != nil {
		return nil, err
	}
	r, ok := data.(files.File)
	if !ok {
		return nil, errors.New("incorrect type from Unixfs().Get()")
	}
	return io.ReadAll(r)
}

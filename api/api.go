package api

import (
	"bytes"
	"context"
	"encoding/json"

	"net/http"
	"os/exec"

	"encoding/base64"

	"github.com/gorilla/mux"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/sunriselayer/sunrise-data/types"
)

type RetriveRequest struct {
	MetadataUri string `json:"metadata_uri"`
}
type RetriveResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type UploadIpfsRequest struct {
	UploadData string `json:"upload_data"`
	Protocol   string `json:"protocol"`
}

type UploadIpfsResponse struct {
	Cid string `json:"cid"`
}
type PublishRequest struct {
	MetadataUri       string   `json:"metadata_uri"`
	ShardDoubleHashes []string `json:"shard_double_hashes"`
}

type PublishResponse struct {
	TxHash string `json:"tx_hash"`
}

func Handle() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to cau-sunrise-data API"))
	})
	r.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		var req PublishRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		addressPrefix := "cosmos"
		client, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Account `alice` was initialized during `ignite chain serve`
		accountName := "alice"
		// Get account from the keyring
		account, err := client.Account(accountName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		addr, err := account.Address(addressPrefix)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		shardDoubleHashes, err := base64StringsToByteSlices(req.ShardDoubleHashes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Define a message to create a post
		msg := &types.MsgPublishData{
			Sender:            addr,
			MetadataUri:       req.MetadataUri,
			ShardDoubleHashes: shardDoubleHashes,
		}
		// Broadcast a transaction from account `alice` with the message
		// to create a post store response in txResp
		txResp, err := client.BroadcastTx(ctx, account, msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Print response from broadcasting a transaction
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PublishResponse{
			TxHash: txResp.TxHash,
		})
	})
	r.HandleFunc("/retrieve", func(w http.ResponseWriter, r *http.Request) {
		var req RetriveRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		command := "sunrised"
		command_arg := "q da show-publish-data " + req.MetadataUri
		cmd := exec.Command(command, command_arg)

		output, err := cmd.Output()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(output)

	}).Methods("POST")

	r.HandleFunc("/uploadIpfs", func(w http.ResponseWriter, r *http.Request) {
		var req UploadIpfsRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		node, err := rpc.NewLocalApi()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx := context.Background()
		uploadDataByteArr, err := base64.StdEncoding.DecodeString(req.UploadData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Create a reader for the byte array
		reader := bytes.NewReader(uploadDataByteArr)
		fileReader := files.NewReaderFile(reader)
		cidFile, err := node.Unixfs().Add(ctx, fileReader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(UploadIpfsResponse{
			Cid: cidFile.String(),
		})
	})
	http.ListenAndServe(":8080", r)
	// http.Handle("/", r)
}

func base64StringsToByteSlices(base64Strings []string) ([][]byte, error) {
	var byteSlices [][]byte

	for _, b64Str := range base64Strings {
		byteSlice, err := base64.StdEncoding.DecodeString(b64Str)
		if err != nil {
			return nil, err
		}
		byteSlices = append(byteSlices, byteSlice)
	}

	return byteSlices, nil
}

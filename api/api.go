package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"net/http"

	"encoding/base64"

	native_mimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/everFinance/goar"
	goar_types "github.com/everFinance/goar/types"
	"github.com/gorilla/mux"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/sunriselayer/sunrise-data/cosmosclient"
	"github.com/sunriselayer/sunrise/x/da/erasurecoding"
	"github.com/sunriselayer/sunrise/x/da/types"
)

const (
	IPFS_PROTOCOL    = "ipfs"
	ARWEAVE_PROTOCOL = "arweave"
)

type UploadedDataResponse struct {
	ShardSize   uint64   `json:"shard_size"`
	ShardCount  uint64   `json:"shard_count"`
	ShardUris   []string `json:"shard_uris"`
	ShardHashes []string `json:"shard_hashes"`
}

type PublishRequest struct {
	Blob           string `json:"blob"`
	ShardCountHalf int    `json:"shard_count_half"`
	Protocol       string `json:"protocol"`
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
		protocol := IPFS_PROTOCOL
		if req.Protocol == IPFS_PROTOCOL {
			protocol = IPFS_PROTOCOL
		} else if req.Protocol == ARWEAVE_PROTOCOL {
			protocol = ARWEAVE_PROTOCOL
		} else {
			http.Error(w, "Invalid Protocol", http.StatusBadRequest)
			return
		}

		recoveredDataHash, err := hashSha256(blobBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		shardSize, shardCount, shards := erasurecoding.ErasureCode(blobBytes, req.ShardCountHalf)
		shardUris, err := getShardUris(shards, protocol)
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
		metadataUri := ""
		if protocol == IPFS_PROTOCOL {
			metadataUri, err = uploadToIpfs(metadataBytes)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			metadataUri, err = uploadToArweave(metadataBytes)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
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
			RecoveredDataHash: recoveredDataHash,
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
		indices := r.URL.Query().Get("indices")
		indicesList := strings.Split(indices, ",")
		if metadataUri == "" {
			http.Error(w, "Invalid query parameter", http.StatusBadRequest)
			return
		}
		metadataBytes, err := getDataFromIpfsOrArweave(metadataUri)
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
		fmt.Println(len(indicesList))
		for _, index := range indicesList {
			iIndex, err := strconv.Atoi(index)
			if err != nil {
				continue
			}
			if iIndex < len(metadata.ShardUris) {
				shardUri := metadata.ShardUris[iIndex]
				shardData, err := getDataFromIpfsOrArweave(shardUri)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				shardHashes = append(shardHashes, base64.StdEncoding.EncodeToString(hashMimc(shardData)))
			}
		}

		res := UploadedDataResponse{
			ShardSize:   metadata.ShardSize,
			ShardCount:  metadata.ShardCount,
			ShardUris:   metadata.ShardUris,
			ShardHashes: shardHashes,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)

	}).Methods("GET")
	fmt.Println("Running Publisher API on localhost:8000")
	http.ListenAndServe(":8000", r)

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

func hashSha256(data []byte) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// upload shards to ipfs
func getShardUris(inputData [][]byte, protocol string) ([]string, error) {
	var shardUris []string
	for _, data := range inputData {
		if protocol == IPFS_PROTOCOL {
			shardUri, err := uploadToIpfs(data)
			if err != nil {
				return nil, err
			}
			shardUris = append(shardUris, shardUri)
		} else {
			shardUri, err := uploadToArweave(data)
			if err != nil {
				return nil, err
			}
			shardUris = append(shardUris, shardUri)
		}
	}
	return shardUris, nil
}

func uploadToIpfs(inputData []byte) (string, error) {
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
func uploadToArweave(inputData []byte) (string, error) {
	wallet, err := goar.NewWalletFromPath("../keyfile.json", "https://arweave.net")
	if err != nil {
		return "", err
	}
	tx, err := wallet.SendData(
		inputData, // Data bytes
		[]goar_types.Tag{},
	)
	if err != nil {
		return "", err
	}
	return tx.ID, nil
}

func getDataFromIpfsOrArweave(uri string) ([]byte, error) {
	if uri[:6] == "/ipfs/" { //ipfs
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
	} else { //arweave
		arweaveClient := goar.NewClient("https://arweave.net")
		return arweaveClient.GetTransactionData(uri)
	}

}

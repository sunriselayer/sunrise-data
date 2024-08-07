package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"net/http"

	native_mimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/everFinance/goar"
	goar_types "github.com/everFinance/goar/types"
	"github.com/gorilla/mux"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/sunriselayer/sunrise-data/config"
)

const (
	IPFS_PROTOCOL    = "ipfs"
	ARWEAVE_PROTOCOL = "arweave"
)

type UploadedDataResponse struct {
	ShardSize   uint64   `json:"shard_size"`
	ShardUris   []string `json:"shard_uris"`
	ShardHashes []string `json:"shard_hashes"`
}

type PublishRequest struct {
	Blob           string `json:"blob"`
	ShardCountHalf int    `json:"shard_count_half"`
	Protocol       string `json:"protocol"`
}

type PublishResponse struct {
	TxHash      string `json:"tx_hash"`
	MetadataUri string `json:"metadata_uri"`
}

type GetBlobResponse struct {
	Blob string `json:"blob"`
}

var Conf config.Config

func Handle(conf config.Config) {
	Conf = conf

	r := mux.NewRouter()
	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to cau-sunrise-data API"))
	}).Methods("GET")
	r.HandleFunc("/api/publish", Publish).Methods("POST")

	r.HandleFunc("/api/uploaded_data", UploadedData).Methods("GET")
	r.HandleFunc("/api/get_blob", GetBlob).Methods("GET")

	fmt.Println("Running Publisher API on localhost:", Conf.Api.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", Conf.Api.Port), r)

}

func byteSlicesToDoubleHashes(inputData [][]byte) [][]byte {
	var convertedData [][]byte
	for _, data := range inputData {
		convertedData = append(convertedData, DoubleHashMimc(data))
	}
	return convertedData
}

func HashMimc(data []byte) []byte {
	m := native_mimc.NewMiMC()
	m.Write(data)
	return m.Sum(nil)
}

func DoubleHashMimc(data []byte) []byte {
	hashData := HashMimc(data)
	return HashMimc(hashData)
}

func HashSha256(data []byte) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// upload shards to ipfs
func GetShardUris(inputData [][]byte, protocol string) ([]string, error) {
	var shardUris []string
	for _, data := range inputData {
		if protocol == IPFS_PROTOCOL {
			shardUri, err := UploadToIpfs(data)
			if err != nil {
				return nil, err
			}
			shardUris = append(shardUris, shardUri)
		} else {
			shardUri, err := UploadToArweave(data)
			if err != nil {
				return nil, err
			}
			shardUris = append(shardUris, shardUri)
		}
	}
	return shardUris, nil
}

func UploadToIpfs(inputData []byte) (string, error) {
	var err error
	var node *rpc.HttpApi
	if Conf.Api.IpfsApiUrl != "" {
		node, err = rpc.NewURLApiWithClient(Conf.Api.IpfsApiUrl, &http.Client{
			Transport: &http.Transport{
				Proxy:             http.ProxyFromEnvironment,
				DisableKeepAlives: true,
			},
		})
	} else {
		node, err = rpc.NewLocalApi()
	}

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
func UploadToArweave(inputData []byte) (string, error) {
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

func GetDataFromIpfsOrArweave(uri string) ([]byte, error) {
	if uri[:6] == "/ipfs/" { //ipfs
		var err error
		var node *rpc.HttpApi
		if Conf.Api.IpfsApiUrl != "" {
			node, err = rpc.NewURLApiWithClient(Conf.Api.IpfsApiUrl, &http.Client{
				Transport: &http.Transport{
					Proxy:             http.ProxyFromEnvironment,
					DisableKeepAlives: true,
				},
			})
		} else {
			node, err = rpc.NewLocalApi()
		}

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

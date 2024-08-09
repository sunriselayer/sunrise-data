package publisher

import (
	"bytes"
	"context"
	"net/http"

	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/kubo/client/rpc"

	"github.com/sunriselayer/sunrise-data/config"
	scontext "github.com/sunriselayer/sunrise-data/context"
)

// upload shards to ipfs
func GetShardUris(inputData [][]byte, protocol string) ([]string, error) {
	var shardUris []string
	for _, data := range inputData {
		if protocol == config.IPFS_PROTOCOL {
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
	if scontext.Config.Api.IpfsApiUrl != "" {
		node, err = rpc.NewURLApiWithClient(scontext.Config.Api.IpfsApiUrl, &http.Client{
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

	return "ipfs://" + cidFile.RootCid().String(), nil
}
func UploadToArweave(inputData []byte) (string, error) {
	wallet, err := goar.NewWalletFromPath("../keyfile.json", "https://arweave.net")
	if err != nil {
		return "", err
	}
	tx, err := wallet.SendData(
		inputData, // Data bytes
		[]types.Tag{},
	)
	if err != nil {
		return "", err
	}
	return "ar://" + tx.ID, nil
}

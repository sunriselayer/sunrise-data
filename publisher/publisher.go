package publisher

import (
	"bytes"
	"context"
	"net/http"
	"sort"

	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/rs/zerolog/log"

	"github.com/sunriselayer/sunrise-data/consts"
	scontext "github.com/sunriselayer/sunrise-data/context"
)

type ShardUriCh struct {
	shardUri string
	index    int
}

// upload shards to ipfs
func GetShardUris(inputData [][]byte, protocol string) ([]string, error) {
	var shardUris []string
	shardUriCh := make(chan ShardUriCh, len(inputData))

	for i, data := range inputData {
		go func() {
			if protocol == consts.IPFS_PROTOCOL {
				shardUri, err := UploadToIpfs(data)
				if err != nil {
					return
				}
				shardUriCh <- ShardUriCh{shardUri, i}
			} else {
				shardUri, err := UploadToArweave(data)
				if err != nil {
					return
				}
				shardUriCh <- ShardUriCh{shardUri, i}
			}
		}()
	}

	var shardUriChs []ShardUriCh
	for range inputData {
		chValue := <-shardUriCh
		shardUriChs = append(shardUriChs, chValue)
	}
	sort.Slice(shardUriChs, func(i, j int) bool {
		return shardUriChs[i].index < shardUriChs[j].index
	})

	for _, shardUriCh := range shardUriChs {
		shardUris = append(shardUris, shardUriCh.shardUri)
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
	reason, isPinned, err := node.Pin().IsPinned(ctx, cidFile)
	log.Info().Msgf("pinned status reason: %s, isPinned: %t, err: %v", reason, isPinned, err)

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

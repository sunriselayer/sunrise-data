package protocols

import (
	"sort"
	"strings"

	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
)

type Arweave struct {
	RpcUrl string
}

func uploadToArweave(inputData []byte) (string, error) {
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

func (arweave *Arweave) PublishShards(shards [][]byte) (uris []string, err error) {
	var shardUris []string
	shardUriCh := make(chan ShardUriCh, len(shards))

	for i, data := range shards {
		go func() {
			shardUri, err := uploadToArweave(data)
			if err != nil {
				return
			}
			shardUriCh <- ShardUriCh{shardUri, i}

		}()
	}

	var shardUriChs []ShardUriCh
	for range shards {
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

func (arweave *Arweave) PublishMetadata(metadata []byte) (uri string, err error) {
	return uploadToArweave(metadata)
}

func (arweave *Arweave) Retrieve(uri string) (shards []byte, err error) {
	arweaveClient := goar.NewClient("https://arweave.net")
	return arweaveClient.GetTransactionData(strings.Replace(uri, "ar://", "", 1))
}

package protocols

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"

	scontext "github.com/sunriselayer/sunrise-data/context"
)

type Ipfs struct {
}

var _ Protocol = &Ipfs{}

func uploadToIpfs(inputData []byte) (string, error) {
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

func (ipfs *Ipfs) PublishShards(shards [][]byte) (uris []string, err error) {
	var shardUris []string
	shardUriCh := make(chan ShardUriCh, len(shards))

	for i, data := range shards {
		go func() {
			shardUri, err := uploadToIpfs(data)
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

func (ipfs *Ipfs) PublishMetadata(metadata []byte) (uri string, err error) {
	return uploadToIpfs(metadata)
}

func (ipfs *Ipfs) Retrieve(uri string) (shards []byte, err error) {
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
		return nil, err
	}
	ctx := context.Background()
	cidData, err := cid.Decode(strings.Replace(uri, "ipfs://", "", 1))
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

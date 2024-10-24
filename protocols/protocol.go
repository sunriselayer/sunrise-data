package protocols

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ipfs/kubo/client/rpc"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sunriselayer/sunrise-data/consts"
	scontext "github.com/sunriselayer/sunrise-data/context"
)

type ShardUriCh struct {
	shardUri string
	index    int
}

type Protocol interface {
	PublishShards(inputData [][]byte) (uris []string, err error)
	PublishMetadata(metadata []byte) (uri string, err error)
	Retrieve(uri string) (shards []byte, err error)
}

func GetPublishProtocol(protocol string) (Protocol, error) {
	if protocol == consts.IPFS_PROTOCOL {
		return &Ipfs{}, nil
	} else if protocol == consts.ARWEAVE_PROTOCOL {
		return &Arweave{}, nil
	}
	return nil, errors.New("unsupported protocol")
}

func GetRetrieveProtocol(uri string) (Protocol, error) {
	if strings.Contains(uri, "ipfs://") {
		return &Ipfs{}, nil
	} else if strings.Contains(uri, "ar://") {
		return &Arweave{}, nil
	}

	return nil, errors.New("unsupported protocol")
}

func ConnectSwarm(addrInfo peer.AddrInfo) error {
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
		return err
	}

	ctx := context.Background()
	err = node.Swarm().Connect(ctx, addrInfo)

	return err
}

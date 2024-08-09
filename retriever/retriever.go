package retriever

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/everFinance/goar"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"

	scontext "github.com/sunriselayer/sunrise-data/context"
)

func GetDataFromIpfsOrArweave(uri string) ([]byte, error) {
	if strings.Contains(uri, "ipfs://") { //ipfs
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
	} else if strings.Contains(uri, "ar://") { //arweave
		arweaveClient := goar.NewClient("https://arweave.net")
		return arweaveClient.GetTransactionData(strings.Replace(uri, "ar://", "", 1))
	} else {
		return nil, errors.New("unsupported protocol")
	}
}

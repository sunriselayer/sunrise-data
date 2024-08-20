package retriever

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/everFinance/goar"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"

	scontext "github.com/sunriselayer/sunrise-data/context"
)

func GetDataFromIpfsOrArweave(uri string) ([]byte, error) {
	fmt.Println("GetDataFromIpfsOrArweave-1", time.Now())
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
		fmt.Println("GetDataFromIpfsOrArweave-2", time.Now())

		if err != nil {
			return nil, err
		}
		fmt.Println("GetDataFromIpfsOrArweave-3", time.Now())
		ctx := context.Background()
		cidData, err := cid.Decode(strings.Replace(uri, "ipfs://", "", 1))
		fmt.Println("GetDataFromIpfsOrArweave-4", time.Now())
		if err != nil {
			return nil, err
		}
		data, err := node.Unixfs().Get(ctx, path.FromCid(cidData))
		fmt.Println("GetDataFromIpfsOrArweave-5", time.Now())
		if err != nil {
			return nil, err
		}
		r, ok := data.(files.File)
		if !ok {
			return nil, errors.New("incorrect type from Unixfs().Get()")
		}
		fmt.Println("GetDataFromIpfsOrArweave-6", time.Now())
		return io.ReadAll(r)
	} else if strings.Contains(uri, "ar://") { //arweave
		fmt.Println("GetDataFromIpfsOrArweave-7", time.Now())
		arweaveClient := goar.NewClient("https://arweave.net")
		return arweaveClient.GetTransactionData(strings.Replace(uri, "ar://", "", 1))
	}

	fmt.Println("GetDataFromIpfsOrArweave-8", time.Now())
	return nil, errors.New("unsupported protocol")
}

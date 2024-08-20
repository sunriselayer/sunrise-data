package retriever

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/everFinance/goar"
	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
	"github.com/rs/zerolog/log"

	scontext "github.com/sunriselayer/sunrise-data/context"
)

func GetDataFromIpfsOrArweave(uri string) ([]byte, error) {
	log.Info().Msgf("GetDataFromIpfsOrArweave-1 %d", time.Now())
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
		log.Info().Msgf("GetDataFromIpfsOrArweave-2 %d", time.Now())

		if err != nil {
			return nil, err
		}
		log.Info().Msgf("GetDataFromIpfsOrArweave-3 %d", time.Now())
		ctx := context.Background()
		cidData, err := cid.Decode(strings.Replace(uri, "ipfs://", "", 1))
		log.Info().Msgf("GetDataFromIpfsOrArweave-4 %d", time.Now())
		if err != nil {
			return nil, err
		}
		data, err := node.Unixfs().Get(ctx, path.FromCid(cidData))
		log.Info().Msgf("GetDataFromIpfsOrArweave-5 %d", time.Now())
		if err != nil {
			return nil, err
		}
		r, ok := data.(files.File)
		if !ok {
			return nil, errors.New("incorrect type from Unixfs().Get()")
		}
		log.Info().Msgf("GetDataFromIpfsOrArweave-6 %d", time.Now())
		return io.ReadAll(r)
	} else if strings.Contains(uri, "ar://") { //arweave
		log.Info().Msgf("GetDataFromIpfsOrArweave-7", time.Now())
		arweaveClient := goar.NewClient("https://arweave.net")
		return arweaveClient.GetTransactionData(strings.Replace(uri, "ar://", "", 1))
	}

	log.Info().Msgf("GetDataFromIpfsOrArweave-8 %d", time.Now())
	return nil, errors.New("unsupported protocol")
}

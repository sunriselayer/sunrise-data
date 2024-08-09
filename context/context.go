package context

import (
	"context"

	datypes "github.com/sunriselayer/sunrise/x/da/types"

	//  "github.com/sunriselayer/sunrise-data/cmd"
	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/cosmosclient"
	"github.com/sunriselayer/sunrise-data/cosmosclient/cosmosaccount"
)

var (
	Ctx         context.Context
	NodeClient  cosmosclient.Client
	QueryClient datypes.QueryClient
	Account     cosmosaccount.Account
	Addr        string
	Config      config.Config
)

func GetContext(conf config.Config) {
	Config = conf
	Ctx = context.Background()
	NodeClient, _ = cosmosclient.New(
		Ctx,
		cosmosclient.WithAddressPrefix(conf.Chain.AddrPrefix),
		cosmosclient.WithKeyringBackend(cosmosaccount.KeyringBackend(conf.Chain.KeyringBackend)),
		cosmosclient.WithHome(conf.Chain.HomePath),
		cosmosclient.WithFees(conf.Chain.Fees),
	)
	QueryClient = datypes.NewQueryClient(NodeClient.Context())

	// Get account from the keyring
	Account, _ = NodeClient.Account(conf.Chain.PublisherAccount)
	Addr, _ = Account.Address(conf.Chain.AddrPrefix)
}

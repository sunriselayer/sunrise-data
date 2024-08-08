package context

import (
	"context"

	//  "github.com/sunriselayer/sunrise-data/cmd"
	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/cosmosclient"
	"github.com/sunriselayer/sunrise-data/cosmosclient/cosmosaccount"
)

var (
	Ctx        context.Context
	NodeClient cosmosclient.Client
	Account    cosmosaccount.Account
	Addr       string
)

func GetContext() {
	Ctx = context.Background()
	NodeClient, _ = cosmosclient.New(
		Ctx,
		cosmosclient.WithAddressPrefix(config.SUNRISE_ADDR_PRIFIX),
		cosmosclient.WithHome(config.SUNRISE_HOME_DIR),
		cosmosclient.WithFees(config.FEES),
		cosmosclient.WithKeyringBackend(config.KEYRING_BACKEND),
	)

	// Get account from the keyring
	Account, _ = NodeClient.Account(config.PUBLISHER_ACCOUNT)
	Addr, _ = Account.Address(config.SUNRISE_ADDR_PRIFIX)
}

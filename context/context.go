package context

import (
	"context"

	datypes "github.com/sunriselayer/sunrise/x/da/types"

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

func GetPublishContext(conf config.Config) error {
	Config = conf
	Ctx = context.Background()
	NodeClient, err := cosmosclient.New(
		Ctx,
		cosmosclient.WithNodeAddress(conf.Chain.SunrisedRPC),
		cosmosclient.WithAddressPrefix(conf.Chain.AddressPrefix),
		cosmosclient.WithKeyringBackend(cosmosaccount.KeyringBackend(conf.Chain.KeyringBackend)),
		cosmosclient.WithHome(conf.Chain.HomePath),
		cosmosclient.WithFees(conf.Publish.PublishFees),
		cosmosclient.WithGasAdjustment(1.5),
		cosmosclient.WithGas(cosmosclient.GasAuto),
	)
	if err != nil {
		return err
	}
	QueryClient = datypes.NewQueryClient(NodeClient.Context())

	// Get publisher account from the keyring
	Account, err = NodeClient.Account(conf.Publish.PublisherAccount)
	if err != nil {
		return err
	}
	Addr, err = Account.Address(conf.Chain.AddressPrefix)
	if err != nil {
		return err
	}
	return nil
}

func GetProofContext(conf config.Config) error {
	Config = conf
	Ctx = context.Background()
	NodeClient, err := cosmosclient.New(
		Ctx,
		cosmosclient.WithNodeAddress(conf.Chain.SunrisedRPC),
		cosmosclient.WithAddressPrefix(conf.Chain.AddressPrefix),
		cosmosclient.WithKeyringBackend(cosmosaccount.KeyringBackend(conf.Chain.KeyringBackend)),
		cosmosclient.WithHome(conf.Chain.HomePath),
		cosmosclient.WithFees(conf.Validator.ProofFees),
		cosmosclient.WithGasAdjustment(1.5),
		cosmosclient.WithGas(cosmosclient.GasAuto),
	)
	if err != nil {
		return err
	}
	QueryClient = datypes.NewQueryClient(NodeClient.Context())

	// Get deputy account from the keyring
	Account, err := NodeClient.Account(conf.Validator.ProofDeputyAccount)
	if err != nil {
		return err
	}
	Addr, err = Account.Address(conf.Chain.AddressPrefix)
	if err != nil {
		return err
	}
	return nil
}

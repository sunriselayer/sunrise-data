package context

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	datypes "github.com/sunriselayer/sunrise/x/da/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

	if conf.Chain.SunrisedRPC == "" {
		return fmt.Errorf("sunrised_rpc is not configured")
	}
	log.Info().Msgf("sunrised_rpc: %s", conf.Chain.SunrisedRPC)

	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(conf.Chain.AddressPrefix, conf.Chain.AddressPrefix+"pub")
	sdkConfig.SetBech32PrefixForValidator(conf.Chain.AddressPrefix+"valoper", conf.Chain.AddressPrefix+"valoperpub")
	sdkConfig.SetBech32PrefixForConsensusNode(conf.Chain.AddressPrefix+"valcons", conf.Chain.AddressPrefix+"valconspub")
	sdkConfig.Seal()

	var err error
	NodeClient, err = cosmosclient.New(
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
		return fmt.Errorf("failed to create cosmos client: %w", err)
	}

	_, err = NodeClient.Status(Ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC at %s: %w", conf.Chain.SunrisedRPC, err)
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
	log.Info().Msgf("publisher address: %v", Addr)
	return nil
}

func GetProofContext(conf config.Config) error {
	Config = conf
	Ctx = context.Background()

	if conf.Chain.SunrisedRPC == "" {
		return fmt.Errorf("sunrised_rpc is not configured")
	}
	log.Info().Msgf("sunrised_rpc: %s", conf.Chain.SunrisedRPC)

	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(conf.Chain.AddressPrefix, conf.Chain.AddressPrefix+"pub")
	sdkConfig.SetBech32PrefixForValidator(conf.Chain.AddressPrefix+"valoper", conf.Chain.AddressPrefix+"valoperpub")
	sdkConfig.SetBech32PrefixForConsensusNode(conf.Chain.AddressPrefix+"valcons", conf.Chain.AddressPrefix+"valconspub")
	sdkConfig.Seal()

	var err error
	NodeClient, err = cosmosclient.New(
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
		return fmt.Errorf("failed to create cosmos client: %w", err)
	}

	_, err = NodeClient.Status(Ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC at %s: %w", conf.Chain.SunrisedRPC, err)
	}

	QueryClient = datypes.NewQueryClient(NodeClient.Context())

	// Get deputy account from the keyring
	Account, err = NodeClient.Account(conf.Validator.ProofDeputyAccount)
	if err != nil {
		return err
	}
	Addr, err = Account.Address(conf.Chain.AddressPrefix)
	if err != nil {
		return err
	}
	log.Info().Msgf("deputy address: %v", Addr)
	return nil
}

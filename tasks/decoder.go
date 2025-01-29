package tasks

import (
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/sunriselayer/sunrise/app"
	"github.com/sunriselayer/sunrise/cmd/sunrised/cmd"
)

func GetTxConfig() client.TxConfig {
	var (
		autoCliOpts   autocli.AppOptions
		moduleManager *module.Manager
		clientCtx     client.Context

		appCodec codec.Codec
	)

	if err := depinject.Inject(
		depinject.Configs(app.AppConfig(),
			depinject.Supply(
				log.NewNopLogger(),
			),
			depinject.Provide(
				cmd.ProvideClientContext,
			),
		),
		&autoCliOpts,
		&moduleManager,
		&clientCtx,
	); err != nil {
		panic(err)
	}

	ibcModules := app.RegisterIBC(clientCtx.InterfaceRegistry, appCodec)
	for name, mod := range ibcModules {
		moduleManager.Modules[name] = module.CoreAppModuleAdaptor(name, mod)
		autoCliOpts.Modules[name] = mod
	}

	return clientCtx.TxConfig
}

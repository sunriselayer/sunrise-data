package cosmosclient

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
)

// signer implements the Signer interface.
type signer struct{}

func (signer) Sign(clientCtx client.Context, txf tx.Factory, name string, txBuilder client.TxBuilder, overwriteSig bool) error {
	return tx.Sign(clientCtx, txf, name, txBuilder, overwriteSig)
}

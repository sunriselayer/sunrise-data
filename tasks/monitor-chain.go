package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cometbft/cometbft/types"

	"github.com/sunriselayer/sunrise-data/config"
)

// MonitorChain
func MonitorChain() {
	fmt.Println("Monitor Chain")
	client, err := http.New(config.TENDERMINT_RPC, "/websocket")
	if err != nil {
		fmt.Print("1", err)
		return
	}

	err = client.Start()
	if err != nil {
		fmt.Print("2", err)
		return
	}

	defer client.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	query := "tm.event = 'Tx'"
	txs, err := client.Subscribe(ctx, "test-client", query)
	if err != nil {
		fmt.Print("3", err)
		return
	}
	fmt.Print("4", txs)

	go func() {
		for e := range txs {
			fmt.Println("got ", e.Data.(types.EventDataTx))
		}
	}()
}

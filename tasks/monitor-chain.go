package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	tmjson "github.com/cometbft/cometbft/libs/json"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	tmTypes "github.com/cometbft/cometbft/rpc/core/types"
	tmJsonRPCTypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	"github.com/cometbft/cometbft/types"

	"github.com/sunriselayer/sunrise-data/config"
)

// MonitorChain
func MonitorChain() {
	fmt.Println("Monitor Chain")
	client, err := rpchttp.New(config.COMETBFT_RPC, "/websocket")
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := "tm.event = 'Tx' AND tx.height = 3"
	txs, err := client.Subscribe(ctx, "test-client", query)
	if err != nil {
		fmt.Print("3", err)
		return
	}
	fmt.Print("4", txs)

	go func() {
		fmt.Print("4", txs)
		for e := range txs {
			fmt.Println("got ", e.Data.(types.EventDataTx))
		}
	}()
}

func SearchTxHashHandle(rpcAddr string, sender string, recipient string, txType string, page int, limit int, txMinHeight int64, txMaxHeight int64, txHash string) (*tmTypes.ResultTxSearch, error) {
	var events = make([]string, 0, 5)

	if sender != "" {
		events = append(events, fmt.Sprintf("transfer.sender='%s'", sender))
	}

	if recipient != "" {
		events = append(events, fmt.Sprintf("transfer.recipient='%s'", recipient))
	}

	if txType != "all" && txType != "" {
		events = append(events, fmt.Sprintf("message.action='%s'", txType))
	}

	if txHash != "" {
		events = append(events, fmt.Sprintf("tx.hash='%s'", txHash))
	}

	if txMinHeight >= 0 {
		events = append(events, fmt.Sprintf("tx.height>=%d", txMinHeight))
	}

	if txMaxHeight >= 0 {
		events = append(events, fmt.Sprintf("tx.height<=%d", txMaxHeight))
	}

	// search transactions
	endpoint := fmt.Sprintf("%s/tx_search?query=\"%s\"&page=%d&&per_page=%d&order_by=\"desc\"", rpcAddr, strings.Join(events, "%20AND%20"), page, limit)
	if page == 0 {
		endpoint = fmt.Sprintf("%s/tx_search?query=\"%s\"&per_page=%d&order_by=\"desc\"", rpcAddr, strings.Join(events, "%20AND%20"), limit)
	}
	fmt.Println("Entering transaction search: ", endpoint)

	resp, err := http.Get(endpoint)
	if err != nil {
		fmt.Println("Unable to connect to ", endpoint)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	response := new(tmJsonRPCTypes.RPCResponse)

	if err := json.Unmarshal(respBody, response); err != nil {
		fmt.Println("Unable to decode response: ", err)
		return nil, err
	}

	if response.Error != nil {
		fmt.Println("Error response:", response.Error.Message)
		return nil, errors.New(response.Error.Message)
	}

	result := new(tmTypes.ResultTxSearch)
	if err := tmjson.Unmarshal(response.Result, result); err != nil {
		fmt.Println("Failed to unmarshal result:", err)
		return nil, fmt.Errorf("error unmarshalling result: %w", err)
	}

	return result, nil
}

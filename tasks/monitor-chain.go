package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	tmjson "github.com/cometbft/cometbft/libs/json"
	tmTypes "github.com/cometbft/cometbft/rpc/core/types"
	tmJsonRPCTypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"

	"github.com/sunriselayer/sunrise-data/config"
)

// MonitorChain
func MonitorChain() {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				currentBlock := GetLatestBlockHeight()
				if currentBlock > latestBlockHeight+config.MONITOR_BLOCK_DELAY {
					result, err := SearchTxHashHandle(config.COMETBFT_RPC, 0, 100, int64(latestBlockHeight))
					if err != nil {
						fmt.Println("Transaction search failed: ", err)
					} else {
						latestBlockHeight += 1
						for _, tx := range result.Txs {
							fmt.Println(tx.TxResult)
							fmt.Println(tx.TxResult.Log)
							// if err != nil {
							// 	fmt.Println("Transaction decode failed: ", err)
							// } else {
							// 	msgs := decoded.GetMsgs()
							// 	for _, msg := range msgs {
							// 		switch msg := msg.(type) {
							// 		case *datypes.MsgPublishData:
							// 			fmt.Println(msg)
							// 			metadataUri := msg.MetadataUri
							// 			// TODO: check metadata uri status
							// 			// TODO: verify metadata hash from ipfs
							// 			SubmitFraudTx(metadataUri)
							// 		}
							// 	}
							// }
						}
					}
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func SearchTxHashHandle(rpcAddr string, page int, limit int, txHeight int64) (*tmTypes.ResultTxSearch, error) {
	var events = make([]string, 0, 5)

	events = append(events, fmt.Sprintf("tx.height=%d", txHeight))
	events = append(events, "message.action='/sunrise.da.MsgPublishData'")

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
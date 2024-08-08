package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/types"
)

var (
	latestBlockHeight = 0
)

// RunTasks is a function to run threads.
func RunTasks() {
	latestBlockHeight = GetLatestBlockHeight()
	fmt.Println("latestBlockHeight ", latestBlockHeight)

	go MonitorChain()
}

// MakeCometbftRPCRequest is a function to make GET request
func MakeCometbftRPCRequest(rpcAddr string, url string, query string, res interface{}) (bool, interface{}, int) {
	endpoint := fmt.Sprintf("%s%s?%s", rpcAddr, url, query)

	resp, err := http.Get(endpoint)
	if err != nil {
		fmt.Println("Unable to connect to ", endpoint)
		return false, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	response := new(types.RPCResponse)
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		fmt.Println("Unable to decode response: : ", err)
		return false, err, http.StatusInternalServerError
	}

	byteData, err := json.Marshal(response.Result)
	if err != nil {
		fmt.Println("Invalid response format", err)
	}

	err = json.Unmarshal(byteData, res)
	if err != nil {
		fmt.Println("Invalid response format", err)
	}

	return true, response.Error, resp.StatusCode
}

func GetLatestBlockHeight() int {
	result := types.ChainStatus{}

	success, _, _ := MakeCometbftRPCRequest(config.COMETBFT_RPC, "/status", "", &result)

	if success {
		blockHeight, err := strconv.Atoi(result.SyncInfo.LatestBlockHeight)
		if err != nil {
			fmt.Println("Invalid response format", err)
		}

		return blockHeight
	}

	return 0
}

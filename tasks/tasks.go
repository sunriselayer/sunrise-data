package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/types"
)

var (
	latestBlockHeight = 0
)

// RunTasks is a function to run threads.
func RunTasks() {
	txConfig := GetTxConfig()
	latestBlockHeight = GetLatestBlockHeight()
	log.Print("latestBlockHeight ", latestBlockHeight)

	go MonitorChain(txConfig)
}

// MakeCometbftRPCRequest is a function to make GET request
func MakeCometbftRPCRequest(rpcAddr string, url string, query string, res interface{}) (bool, interface{}, int) {
	endpoint := fmt.Sprintf("%s%s?%s", rpcAddr, url, query)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Print("Unable to connect to ", endpoint)
		return false, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	response := new(types.RPCResponse)
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Print("Unable to decode response: : ", err)
		return false, err, http.StatusInternalServerError
	}

	byteData, err := json.Marshal(response.Result)
	if err != nil {
		log.Print("Invalid response format", err)
	}

	err = json.Unmarshal(byteData, res)
	if err != nil {
		log.Print("Invalid response format", err)
	}

	return true, response.Error, resp.StatusCode
}

func GetLatestBlockHeight() int {
	result := types.ChainStatus{}

	success, _, _ := MakeCometbftRPCRequest(context.Config.Chain.CometbftRPC, "/status", "", &result)

	if success {
		blockHeight, err := strconv.Atoi(result.SyncInfo.LatestBlockHeight)
		if err != nil {
			log.Print("Invalid response format", err)
		}

		return blockHeight
	}

	return 0
}

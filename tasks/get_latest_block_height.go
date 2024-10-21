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

// MakeCometbftRPCRequest is a function to make GET request
func MakeCometbftRPCRequest(rpcAddr string, url string, query string, res interface{}) (bool, interface{}, int) {
	endpoint := fmt.Sprintf("%s%s?%s", rpcAddr, url, query)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Error().Msgf("Unable to connect to %s", endpoint)
		return false, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	response := new(types.RPCResponse)
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Error().Msgf("Unable to decode response: %s", err)
		return false, err, http.StatusInternalServerError
	}

	byteData, err := json.Marshal(response.Result)
	if err != nil {
		log.Error().Msgf("Invalid response format %s", err)
	}

	err = json.Unmarshal(byteData, res)
	if err != nil {
		log.Error().Msgf("Invalid response format %s", err)
	}

	return true, response.Error, resp.StatusCode
}

func GetLatestBlockHeight() int {
	result := types.ChainStatus{}

	success, _, _ := MakeCometbftRPCRequest(context.Config.Chain.CometbftRPC, "/status", "", &result)
	if !success {
		return 0
	}

	blockHeight, err := strconv.Atoi(result.SyncInfo.LatestBlockHeight)
	if err != nil {
		log.Error().Msgf("Invalid response format %s", err)
	}

	return blockHeight
}

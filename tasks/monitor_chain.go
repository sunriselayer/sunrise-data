package tasks

import (
	"encoding/base64"
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
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
	datypes "github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/protocols"
	"github.com/sunriselayer/sunrise-data/publisher"
	"github.com/sunriselayer/sunrise-data/utils"
)

func MonitorChain(txConfig client.TxConfig) {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				currentBlock := GetLatestBlockHeight()
				if currentBlock > latestBlockHeight+context.Config.Chain.VoteExtensionPeriod {
					MonitorBlock(txConfig, int64(latestBlockHeight))
					latestBlockHeight += 1
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func MonitorBlock(txConfig client.TxConfig, syncBlock int64) {
	result, err := SearchTxHashHandle(context.Config.Chain.CometbftRPC, 0, 100, syncBlock)
	if err != nil {
		log.Error().Msgf("Transaction search failed: %s", err)
		return
	}

	for _, tx := range result.Txs {
		decoded, err := txConfig.TxDecoder()(tx.Tx)
		if err != nil {
			log.Error().Msgf("Transaction decode failed: %s", err)
			continue
		}

		msgs := decoded.GetMsgs()
		for _, msg := range msgs {
			switch msg := msg.(type) {
			case *datypes.MsgPublishData:
				metadataUri := msg.MetadataUri

				// For testing fraud TX submission
				// SubmitFraudTx(metadataUri)
				// continue

				// check metadata uri status
				publishedDataResponse, err := context.QueryClient.PublishedData(context.Ctx, &datypes.QueryPublishedDataRequest{MetadataUri: metadataUri})
				if err != nil {
					log.Error().Msgf("Failed to query metadata from on-chain: %s", err)
					continue
				}

				publishedData := publishedDataResponse.Data
				if publishedData.Status != "vote_extension" {
					log.Error().Msg("Not passed the vote extension yet")
					continue
				}

				peerAddrInfo, err := peer.AddrInfoFromString(publishedData.DataSourceInfo)
				if err == nil {
					publisher.ConnectSwarm(*peerAddrInfo)
				}

				// verify shard data
				protocol, err := protocols.GetRetrieveProtocol(metadataUri)
				if err != nil {
					log.Error().Msgf("Failed to get protocol: %s", err)
					SubmitFraudTx(metadataUri)
					continue
				}

				metadataBytes, err := protocol.Retrieve(metadataUri)
				if err != nil {
					log.Error().Msgf("Failed to get metadata: %s", err)
					SubmitFraudTx(metadataUri)
					continue
				}

				metadata := datypes.Metadata{}
				if err := metadata.Unmarshal(metadataBytes); err != nil {
					log.Error().Msgf("Failed to decode metadata: %s", err)
					SubmitFraudTx(metadataUri)
					continue
				}

				if len(publishedData.ShardDoubleHashes) != len(metadata.ShardUris) {
					log.Error().Msgf("Incorrect shard data count: %s", err)
					SubmitFraudTx(metadataUri)
					continue
				}

				validShards := [][]byte{}

				for index, doubleHash := range publishedData.ShardDoubleHashes {
					shardUri := metadata.ShardUris[index]
					shardData, err := protocol.Retrieve(shardUri)
					if err != nil {
						log.Error().Msgf("Failed to get shard data: %s", err)
						continue
					}

					doubleShardHash := base64.StdEncoding.EncodeToString(utils.DoubleHashMimc(shardData))
					if doubleShardHash != base64.StdEncoding.EncodeToString(doubleHash) {
						log.Error().Msgf("Incorrect shard data: %d", index)
						continue
					}
					validShards = append(validShards, shardData)
				}

				DataShardCount := len(publishedData.ShardDoubleHashes) - int(metadata.ParityShardCount)

				if len(validShards) < DataShardCount {
					log.Error().Msgf("Valid shard count less than DataShardCount: %d", len(validShards))
					SubmitFraudTx(metadataUri)
					continue
				}
			}
		}
	}
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
	log.Print("Entering transaction search: ", endpoint)

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Error().Msgf("Unable to connect to %s", endpoint)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	response := new(tmJsonRPCTypes.RPCResponse)

	if err := json.Unmarshal(respBody, response); err != nil {
		log.Error().Msgf("Unable to decode response: %s", err)
		return nil, err
	}

	if response.Error != nil {
		log.Error().Msgf("Error response: %s", response.Error.Message)
		return nil, errors.New(response.Error.Message)
	}

	result := new(tmTypes.ResultTxSearch)
	if err := tmjson.Unmarshal(response.Result, result); err != nil {
		log.Error().Msgf("Failed to unmarshal result: %s", err)
		return nil, fmt.Errorf("error unmarshalling result: %w", err)
	}

	return result, nil
}

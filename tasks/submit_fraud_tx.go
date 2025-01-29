package tasks

import (
	"bufio"
	"bytes"
	"encoding/base64"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
	datypes "github.com/sunriselayer/sunrise/x/da/types"
	"github.com/sunriselayer/sunrise/x/da/zkp"

	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/protocols"
	"github.com/sunriselayer/sunrise-data/utils"
)

func getShardProofBytes(shardHash []byte, shardDoubleHash []byte) ([]byte, bool) {
	queryParamsResponse, err := context.QueryClient.Params(context.Ctx, &datypes.QueryParamsRequest{})
	if err != nil {
		return nil, false
	}

	params := queryParamsResponse.Params

	// witness definition
	assignment := zkp.ValidityProofCircuit{
		ShardHash:       shardHash,
		ShardDoubleHash: shardDoubleHash,
	}

	// compiles our circuit into a R1CS
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &zkp.ValidityProofCircuit{})
	if err != nil {
		return nil, false
	}

	// Recover proving key
	provingKey, err := zkp.UnmarshalProvingKey(params.ZkpProvingKey)
	if err != nil {
		return nil, false
	}

	witness1, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, false
	}

	// groth16: Prove & Verify
	proof, err := groth16.Prove(ccs, provingKey, witness1)
	if err != nil {
		return nil, false
	}

	var b bytes.Buffer
	bufWrite := bufio.NewWriter(&b)
	_, err = proof.WriteTo(bufWrite)
	if err != nil {
		return nil, false
	}

	err = bufWrite.Flush()
	if err != nil {
		return nil, false
	}

	proofBytes := b.Bytes()

	return proofBytes, true
}

// Delete this function
// func submitInvalidDataTx(metadataUri string) bool {
// 	proofMsg := &datypes.MsgSubmitProof{
// 		Sender:      context.Addr,
// 		MetadataUri: metadataUri,
// 		Indices:     []int64{},
// 		Proofs:      [][]byte{},
// 		IsValidData: false,
// 	}

// 	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, proofMsg)
// 	if err != nil {
// 		log.Error().Msgf("Failed to broadcast MsgSubmitProof transaction: %s %s", metadataUri, err)
// 		return false
// 	}
// 	log.Info().Msgf("MsgSubmitProof TxHash: %s", txResp.TxHash)

// 	return true
// }

func submitValidDataTx(metadataUri string, indices []int64, proofs [][]byte) bool {
	proofMsg := &datypes.MsgSubmitProof{
		Sender:      context.Addr,
		MetadataUri: metadataUri,
		Indices:     indices,
		Proofs:      proofs,
	}

	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, proofMsg)
	if err != nil {
		log.Error().Msgf("Failed to broadcast MsgSubmitProof transaction: %s %s", metadataUri, err)
		return false
	}
	log.Info().Msgf("MsgSubmitProof TxHash: %s", txResp.TxHash)

	return true
}

func submitChallengeForFraud(metadataUri string) bool {
	msg := &datypes.MsgChallengeForFraud{
		Sender:      context.Addr,
		MetadataUri: metadataUri,
	}
	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, msg)
	if err != nil {
		log.Error().Msgf("Failed to broadcast MsgChallengeForFraud transaction: %s %s", metadataUri, err)
		return false
	}
	log.Info().Msgf("ChallengeForFraud TxHash: %s", txResp.TxHash)
	return true
}

func SubmitFraudTx(metadataUri string) bool {
	publishedDataResponse, err := context.QueryClient.PublishedData(context.Ctx, &datypes.QueryPublishedDataRequest{MetadataUri: metadataUri})
	if err != nil {
		log.Error().Msgf("Failed to query metadata from on-chain: %s", err)
		return false
	}
	publishedData := publishedDataResponse.Data

	peerAddrInfo, err := peer.AddrInfoFromString(publishedData.DataSourceInfo)
	if err == nil {
		protocols.ConnectSwarm(*peerAddrInfo)
	}

	if context.Config.Api.SubmitChallenge {
		ok := submitChallengeForFraud(metadataUri)
		if !ok {
			return false
		}
	}

	if !context.Config.Api.SubmitProof {
		return false
	}

	protocol, err := protocols.GetRetrieveProtocol(metadataUri)
	if err != nil {
		log.Error().Msgf("Failed to get protocol: %s", err)
		return false
	}

	// verify shard data
	metadataBytes, err := protocol.Retrieve(metadataUri)
	if err != nil {
		log.Error().Msgf("Failed to get metadata: %s", err)
		return false
	}

	metadata := datypes.Metadata{}
	if err := metadata.Unmarshal(metadataBytes); err != nil {
		log.Error().Msgf("Failed to decode metadata: %s", err)
		return false
	}

	if len(publishedData.ShardDoubleHashes) != len(metadata.ShardUris) {
		log.Error().Msgf("Incorrect shard data count: %d %d", len(publishedData.ShardDoubleHashes), len(metadata.ShardUris))
		return false
	}

	validShards := [][]byte{}
	validShardIndexes := []int{}

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
		validShardIndexes = append(validShardIndexes, index)
	}

	DataShardCount := len(publishedData.ShardDoubleHashes) - int(metadata.ParityShardCount)

	if len(validShards) < DataShardCount {
		log.Error().Msgf("Valid shard count less than DataShardCount: %d", len(validShards))
		return false
	}

	shardLength := len(metadata.ShardUris)
	queryThresholdResponse, err := context.QueryClient.ZkpProofThreshold(context.Ctx, &datypes.QueryZkpProofThresholdRequest{ShardCount: uint64(shardLength)})
	if err != nil {
		log.Error().Msgf("Failed to query Threshold: %s", err)
		return false
	}

	threshold := queryThresholdResponse.Threshold
	addr, err := sdk.AccAddressFromBech32(context.Addr)
	if err != nil {
		log.Error().Msgf("Failed to parse AccAddress: %s %s", context.Addr, err)
		return false
	}

	requiredIndices := datypes.ShardIndicesForValidator(sdk.ValAddress(addr), int64(threshold), int64(shardLength))
	proofs := [][]byte{}
	indices := []int64{}

	for _, indice := range requiredIndices {
		for i, validIndex := range validShardIndexes {
			if indice != int64(validIndex) {
				continue
			}

			shardData := validShards[i]
			shardHash := utils.HashMimc(shardData)
			doubleShardHash := utils.HashMimc(shardHash)
			proofBytes, ok := getShardProofBytes(shardHash, doubleShardHash)
			if !ok {
				log.Error().Msgf("Failed to generate shard proof: %s, indice: %d", metadataUri, indice)
				return false
			}

			proofs = append(proofs, proofBytes)
			indices = append(indices, indice)
		}
	}

	return submitValidDataTx(metadataUri, indices, proofs)
}

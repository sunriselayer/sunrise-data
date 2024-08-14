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
	"github.com/rs/zerolog/log"
	datypes "github.com/sunriselayer/sunrise/x/da/types"
	"github.com/sunriselayer/sunrise/x/da/zkp"

	"github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/retriever"
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

func submitInvalidDataTx(metadataUri string) bool {
	proofMsg := &datypes.MsgSubmitProof{
		Sender:      context.Addr,
		MetadataUri: metadataUri,
		Indices:     []int64{},
		Proofs:      [][]byte{},
		IsValidData: false,
	}

	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, proofMsg)
	if err != nil {
		log.Print("Failed to broadcast MsgSubmitProof transaction: ", metadataUri, err)
		return false
	}
	log.Print("MsgSubmitProof TxHash:", txResp.TxHash)

	return true
}

func submitValidDataTx(metadataUri string, indices []int64, proofs [][]byte) bool {
	proofMsg := &datypes.MsgSubmitProof{
		Sender:      context.Addr,
		MetadataUri: metadataUri,
		Indices:     indices,
		Proofs:      proofs,
		IsValidData: true,
	}

	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, proofMsg)
	if err != nil {
		log.Print("Failed to broadcast MsgSubmitProof transaction: ", metadataUri, err)
		return false
	}
	log.Print("MsgSubmitProof TxHash:", txResp.TxHash)

	return true
}

func SubmitFraudTx(metadataUri string) bool {
	msg := &datypes.MsgChallengeForFraud{
		Sender:      context.Addr,
		MetadataUri: metadataUri,
	}
	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, msg)
	if err != nil {
		log.Print("Failed to broadcast MsgChallengeForFraud transaction: ", metadataUri, err)
		return false
	}
	log.Print("ChallengeForFraud TxHash:", txResp.TxHash)

	publishedDataResponse, err := context.QueryClient.PublishedData(context.Ctx, &datypes.QueryPublishedDataRequest{MetadataUri: metadataUri})
	if err != nil {
		log.Print("Failed to query metadata from on-chain: ", err)
		return false
	}
	publishedData := publishedDataResponse.Data

	// verify shard data
	metadataBytes, err := retriever.GetDataFromIpfsOrArweave(metadataUri)
	if err != nil {
		log.Print("Failed to get metadata: ", err)
		return submitInvalidDataTx(metadataUri)
	}

	metadata := datypes.Metadata{}
	if err := metadata.Unmarshal(metadataBytes); err != nil {
		log.Print("Failed to decode metadata: ", err)
		return submitInvalidDataTx(metadataUri)
	}

	if len(publishedData.ShardDoubleHashes) != len(metadata.ShardUris) {
		log.Print("Incorrect shard data count: ", len(publishedData.ShardDoubleHashes), len(metadata.ShardUris))
		return submitInvalidDataTx(metadataUri)
	}

	for index := 0; index < len(metadata.ShardUris); index++ {
		shardUri := metadata.ShardUris[index]
		shardData, err := retriever.GetDataFromIpfsOrArweave(shardUri)
		if err != nil {
			log.Print("Failed to get shard data: ", err)
			return submitInvalidDataTx(metadataUri)
		}
		doubleShardHash := base64.StdEncoding.EncodeToString(utils.DoubleHashMimc(shardData))

		if doubleShardHash != base64.StdEncoding.EncodeToString(publishedData.ShardDoubleHashes[index]) {
			log.Print("Incorrect shard data: ", index)
			return submitInvalidDataTx(metadataUri)
		}
	}

	queryThresholdResponse, err := context.QueryClient.ZkpProofThreshold(context.Ctx, &datypes.QueryZkpProofThresholdRequest{ShardCount: uint64(len(metadata.ShardUris))})
	if err != nil {
		log.Print("Failed to query Threshold: ", err)
		return false
	}

	threshold := queryThresholdResponse.Threshold

	shardLength := int64(len(metadata.ShardUris))
	addr, err := sdk.AccAddressFromBech32(context.Addr)
	if err != nil {
		log.Print("Failed to parse AccAddress: ", context.Addr, err)
		return false
	}
	indices := datypes.ShardIndicesForValidator(sdk.ValAddress(addr), int64(threshold), shardLength)

	proofs := [][]byte{}

	for i := range indices {
		shardUri := metadata.ShardUris[indices[i]]
		shardData, _ := retriever.GetDataFromIpfsOrArweave(shardUri)
		shardHash := utils.HashMimc(shardData)
		doubleShardHash := utils.HashMimc(shardHash)

		proofBytes, ok := getShardProofBytes(shardHash, doubleShardHash)
		if !ok {
			log.Print("Failed to generate shard proof: ", metadataUri, "indice: ", indices[i])
			return false
		}

		proofs = append(proofs, proofBytes)
	}

	return submitValidDataTx(metadataUri, indices, proofs)
}

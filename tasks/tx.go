package tasks

import (
	"bufio"
	"bytes"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/rs/zerolog/log"
	datypes "github.com/sunriselayer/sunrise/x/da/types"
	"github.com/sunriselayer/sunrise/x/da/zkp"

	"github.com/sunriselayer/sunrise-data/context"
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

func submitValidityProof(metadataUri string, indices []int64, proofs [][]byte) bool {
	proofMsg := &datypes.MsgSubmitValidityProof{
		Sender:           context.Addr,
		ValidatorAddress: context.Config.Validator.ValidatorAddress,
		MetadataUri:      metadataUri,
		Indices:          indices,
		Proofs:           proofs,
	}

	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, proofMsg)
	if err != nil {
		log.Error().Msgf("Failed to broadcast MsgSubmitValidityProof transaction: %s %s", metadataUri, err)
		return false
	}
	log.Info().Msgf("MsgSubmitValidityProof TxHash: %s", txResp.TxHash)

	return true
}

func submitInvalidity(metadataUri string, indices []int64) bool {
	msg := &datypes.MsgSubmitInvalidity{
		Sender:      context.Addr,
		MetadataUri: metadataUri,
		Indices:     indices,
	}
	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, msg)
	if err != nil {
		log.Error().Msgf("Failed to broadcast MsgSubmitInvalidity transaction: %s %s", metadataUri, err)
		return false
	}
	log.Info().Msgf("MsgSubmitInvalidity TxHash: %s", txResp.TxHash)
	return true
}

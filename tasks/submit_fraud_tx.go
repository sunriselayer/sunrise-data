package tasks

import (
	"github.com/rs/zerolog/log"
	"github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/context"
)

func SubmitFraudTx(metadataUri string) bool {
	msg := &types.MsgChallengeForFraud{
		Sender:      context.Addr,
		MetadataUri: metadataUri,
	}
	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, msg)
	if err != nil {
		log.Print("Failed to broadcast MsgChallengeForFraud transaction: ", metadataUri, err)
		return false
	}
	log.Print("ChallengeForFraud TxHash:", txResp.TxHash)

	proofMsg := &types.MsgSubmitProof{
		Sender:      context.Addr,
		MetadataUri: metadataUri,
		Indices:     []uint64{},
		ShardHashes: [][]byte{},
	}
	txResp, err = context.NodeClient.BroadcastTx(context.Ctx, context.Account, proofMsg)
	if err != nil {
		log.Print("Failed to broadcast MsgSubmitProof transaction: ", metadataUri, err)
		return false
	}
	log.Print("MsgSubmitProof TxHash:", txResp.TxHash)

	return true
}

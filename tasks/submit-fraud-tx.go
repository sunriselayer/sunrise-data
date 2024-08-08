package tasks

import (
	"fmt"

	"github.com/sunriselayer/sunrise/x/da/types"

	"github.com/sunriselayer/sunrise-data/context"
)

func SubmitFraudTx(metadataUri string) bool {
	// Define a message to create a post
	msg := &types.MsgChallengeForFraud{
		Sender:      context.Addr,
		MetadataUri: metadataUri,
	}
	// Broadcast a transaction from account `alice` with the message
	// to create a post store response in txResp
	txResp, err := context.NodeClient.BroadcastTx(context.Ctx, context.Account, msg)
	if err != nil {
		fmt.Println("Failed to broadcast MsgChallengeForFraud transaction: ", metadataUri, err)
		return false
	}
	fmt.Println("TxHash:", txResp.TxHash)

	return true
}

package tasks

import (
	"github.com/rs/zerolog/log"
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

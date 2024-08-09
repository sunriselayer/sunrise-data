package types

type SyncInfo struct {
	LatestBlockHeight string `json:"latest_block_height,omitempty"`
}

// Used to parse response from tendermint api ("/status")
type ChainStatus struct {
	SyncInfo SyncInfo `json:"sync_info,omitempty"`
}

// RPCResponse is a struct of RPC response
type RPCResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

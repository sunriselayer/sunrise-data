package protocols

type Arweave struct {
	RpcUrl string
}

func (arweave *Arweave) Publish(shards [][]byte) (uris []string) {
	return
}

func (arweave *Arweave) Retrieve(uris map[int]string) (shards map[int]byte) {
	return
}

package protocols

type Protocol interface {
	Publish(shards [][]byte) (uris []string)
	Retrieve(uris map[int]string) (shards map[int]byte)
}

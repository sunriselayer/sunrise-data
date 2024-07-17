package protocols

import (
	"context"

	path "github.com/ipfs/boxo/path"
	"github.com/ipfs/kubo/client/rpc"
)

type Ipfs struct {
}

func (ipfs *Ipfs) Publish(shards [][]byte) (uris []string) {
	return
}

func (ipfs *Ipfs) Retrieve(uris map[int]string) (shards map[int]byte) {
	return
}

func dummy() {
	// "Connect" to local node
	node, err := rpc.NewLocalApi()
	if err != nil {
		panic(err)
	}
	// Pin a given file by its CID
	ctx := context.Background()
	cid := "bafkreidtuosuw37f5xmn65b3ksdiikajy7pwjjslzj2lxxz2vc4wdy3zku"
	p, err := path.NewPath(cid)
	if err != nil {
		panic(err)
	}
	err = node.Pin().Add(ctx, p)
	if err != nil {
		panic(err)
	}
	return
}

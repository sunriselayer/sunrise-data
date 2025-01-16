package rollkit

import (
	"context"
	"encoding/base64"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rollkit/go-da"
	"github.com/sunriselayer/sunrise-data/api"
	"github.com/sunriselayer/sunrise-data/config"
)

type SunriseDA struct {
	ctx    context.Context
	config config.Config
}

func NewSunriseDA(ctx context.Context, config config.Config) SunriseDA {
	return SunriseDA{
		ctx:    ctx,
		config: config,
	}
}

var _ da.DA = &SunriseDA{}

func (sunrise *SunriseDA) MaxBlobSize(ctx context.Context) (uint64, error) {
	var maxBlobSize uint64 = 64 * 64 * 500
	return maxBlobSize, nil
}

func (sunrise *SunriseDA) Get(ctx context.Context, ids []da.ID, namespace da.Namespace) ([]da.Blob, error) {
	var blobs []da.Blob

	return blobs, nil
}

func (c *SunriseDA) GetIDs(ctx context.Context, height uint64, namespace da.Namespace) (*da.GetIDsResult, error) {
	heightAsUint32 := uint32(height)
	ids := make([]byte, 8)
	binary.BigEndian.PutUint32(ids, heightAsUint32)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &da.GetIDsResult{
		IDs:       [][]byte{ids},
		Timestamp: sdkCtx.BlockTime(),
	}, nil
}

func (sunrise *SunriseDA) GetProofs(ctx context.Context, ids []da.ID, namespace da.Namespace) ([]da.Proof, error) {
	var proofs []da.Proof

	return proofs, nil
}

func (sunrise *SunriseDA) Commit(ctx context.Context, daBlobs []da.Blob, namespace da.Namespace) ([]da.Commitment, error) {
	var commitments []da.Commitment

	return commitments, nil
}

func (sunrise *SunriseDA) Submit(ctx context.Context, daBlobs []da.Blob, gasPrice float64, namespace da.Namespace) ([]da.ID, error) {
	var ids []da.ID
	for _, blob := range daBlobs {
		encodedBlob := base64.StdEncoding.EncodeToString(blob)
		req := api.PublishRequest{
			Blob:             encodedBlob,
			DataShardCount:   int(sunrise.config.Rollkit.DataShardCount),
			ParityShardCount: int(sunrise.config.Rollkit.ParityShardCount),
			Protocol:         "ipfs",
		}
		res, err := api.PublishData(req)
		if err != nil {
			return nil, err
		}
		ids = append(ids, []byte(res.MetadataUri))
	}

	return ids, nil
}

func (sunrise *SunriseDA) SubmitWithOptions(ctx context.Context, daBlobs []da.Blob, gasPrice float64, namespace da.Namespace, options []byte) ([]da.ID, error) {
	var ids []da.ID

	return ids, nil
}

func (sunrise *SunriseDA) Validate(ctx context.Context, ids []da.ID, daProofs []da.Proof, namespace da.Namespace) ([]bool, error) {
	var valid []bool

	return valid, nil
}

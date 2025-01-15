package rollkit

import (
	"context"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rollkit/go-da"
)

type SunriseDA struct {
	ctx context.Context
}

func NewSunriseDA(ctx context.Context) SunriseDA {
	return SunriseDA{
		ctx: ctx,
	}
}

var _ da.DA = &SunriseDA{}

func (sunrise *SunriseDA) MaxBlobSize(ctx context.Context) (uint64, error) {
	var maxBlobSize uint64 = 0
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

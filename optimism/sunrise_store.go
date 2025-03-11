package optimism

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	altda "github.com/ethereum-optimism/optimism/op-alt-da"
	"github.com/ethereum/go-ethereum/log"
	api "github.com/sunriselayer/sunrise-data/api"
)

const VersionByte = 0x0c

type SunriseConfig struct {
	URL              string
	DataShardCount   int
	ParityShardCount int
}

// SunriseStore implements DAStorage with sunrise-data backend
type SunriseStore struct {
	Log        log.Logger
	Config     SunriseConfig
	GetTimeout time.Duration
	Namespace  []byte
}

// NewSunriseStore returns a sunrise store.
func NewSunriseStore(cfg SunriseConfig, log log.Logger) *SunriseStore {
	return &SunriseStore{
		Log:        log,
		Config:     cfg,
		GetTimeout: time.Minute,
	}
}

func (d *SunriseStore) Get(ctx context.Context, comm []byte) ([]byte, error) {
	commExtracted, err := Decode(comm)
	if err != nil {
		return nil, fmt.Errorf("sunrise-alt-da: failed to decode payload: %w", err)
	}

	metadataUri := string(commExtracted)
	d.Log.Info("sunrise-alt-da: blob request", "id", metadataUri)

	_, cancel := context.WithTimeout(context.Background(), d.GetTimeout)
	res, err := api.GetBlobData(metadataUri)
	cancel()

	if err != nil {
		return nil, fmt.Errorf("sunrise-alt-da: failed to get blob: %w", err)
	}

	blobs := res.Blob

	if len(blobs) == 0 {
		return nil, fmt.Errorf("sunrise-alt-da: failed to resolve frame: %w", err)
	}

	input, err := base64.StdEncoding.DecodeString(blobs)
	if err != nil {
		return nil, fmt.Errorf("sunrise-alt-da: failed to decode base64: %w", err)
	}

	return input, nil
}

func (d *SunriseStore) Put(ctx context.Context, data []byte) ([]byte, error) {
	req := api.PublishRequest{
		Blob:             base64.StdEncoding.EncodeToString(data),
		DataShardCount:   d.Config.DataShardCount,
		ParityShardCount: d.Config.ParityShardCount,
		Protocol:         "ipfs",
	}

	res, err := api.PublishData(req)
	if err != nil {
		return nil, fmt.Errorf("sunrise-alt-da: failed to post publish request: %w", err)
	}

	d.Log.Info("sunrise-alt-da: blob successfully submitted", "tx_hash", res.TxHash, "uri", res.MetadataUri)
	commitment := altda.NewGenericCommitment(append([]byte{VersionByte}, []byte(res.MetadataUri)...))

	return commitment.Encode(), nil
}

func Decode(comm []byte) ([]byte, error) {
	if comm[0] != 0x01 && comm[1] != VersionByte {
		return nil, fmt.Errorf("invalid encoding")
	}
	return comm[2:], nil
}

package sunrise

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	resp, err := http.Get(fmt.Sprintf("%s/api/get-blob?metadata_uri=%s", d.Config.URL, metadataUri))
	cancel()

	if err != nil {
		return nil, fmt.Errorf("sunrise-alt-da: failed to get blob: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("sunrise-alt-da: failed to read response body: %w", err)
	}

	blobResp := api.GetBlobResponse{}
	err = json.Unmarshal(body, &blobResp)
	if err != nil {
		return nil, fmt.Errorf("sunrise-alt-da: failed to unmarshal response body: %w", err)
	}
	blobs := blobResp.Blob

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
	publishReq := api.PublishRequest{
		Blob:             base64.StdEncoding.EncodeToString(data),
		DataShardCount:   d.Config.DataShardCount,
		ParityShardCount: d.Config.ParityShardCount,
		Protocol:         "ipfs",
	}
	jsonData, err := json.Marshal(publishReq)
	if err != nil {
		return nil, fmt.Errorf("sunrise-alt-da: failed to marshal publish request: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/publish", d.Config.URL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("sunrise-alt-da: failed to post publish request: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("sunrise-alt-da: failed to read response body: %w", err)
	}

	publishResp := api.PublishResponse{}
	err = json.Unmarshal(body, &publishResp)
	if err != nil {
		return nil, fmt.Errorf("sunrise-alt-da: failed to unmarshal response body: %w", err)
	}

	d.Log.Info("sunrise-alt-da: blob successfully submitted", "tx_hash", publishResp.TxHash, "uri", publishResp.MetadataUri)
	commitment := altda.NewGenericCommitment(append([]byte{VersionByte}, []byte(publishResp.MetadataUri)...))

	return commitment.Encode(), nil
}

func Decode(comm []byte) ([]byte, error) {
	if comm[0] != 0x01 && comm[1] != VersionByte {
		return nil, fmt.Errorf("invalid encoding")
	}
	return comm[2:], nil
}

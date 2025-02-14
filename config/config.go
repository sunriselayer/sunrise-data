package config

import (
	toml "github.com/pelletier/go-toml"
)

type Config struct {
	Api struct {
		Port            int    `toml:"port"`
		IpfsApiUrl      string `toml:"ipfs_api_url"`
		IpfsAddrInfo    string `toml:"ipfs_addrinfo"`
		SubmitChallenge bool   `toml:"submit_challenge"`
		SubmitProof     bool   `toml:"submit_proof"`
	}
	Chain struct {
		AddrPrefix             string `toml:"addr_prefix"`
		PublisherAccount       string `toml:"publisher_account"`
		ProofDeputyAccount     string `toml:"proof_deputy_account"`
		ValidatorAddress       string `toml:"validator_address"`
		HomePath               string `toml:"home_path"`
		KeyringBackend         string `toml:"keyring_backend"`
		Fees                   string `toml:"fees"`
		SunrisedRPC            string `toml:"sunrised_rpc"`
		ValidatorProofInterval int    `toml:"validator_proof_interval"`
	}
	Rollkit struct {
		DataShardCount   int `toml:"data_shard_count"`
		ParityShardCount int `toml:"parity_shard_count"`
		Port             int `toml:"port"`
	}
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	configTree, err := toml.LoadFile("config.toml")
	if err != nil {
		return nil, err
	}
	err = configTree.Unmarshal(config)
	return config, err
}

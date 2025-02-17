package config

import (
	toml "github.com/pelletier/go-toml"
)

type Config struct {
	Api struct {
		Port            int    `toml:"port"`
		IpfsApiUrl      string `toml:"ipfs_api_url"`
		IpfsAddressInfo string `toml:"ipfs_address_info"`
	}
	Chain struct {
		AddressPrefix  string `toml:"address_prefix"`
		HomePath       string `toml:"home_path"`
		KeyringBackend string `toml:"keyring_backend"`
		SunrisedRPC    string `toml:"sunrised_rpc"`
	}
	Publish struct {
		PublisherAccount string `toml:"publisher_account"`
		PublishFees      string `toml:"publish_fees"`
	}
	Validator struct {
		ProofDeputyAccount string `toml:"proof_deputy_account"`
		ValidatorAddress   string `toml:"validator_address"`
		ProofFees          string `toml:"proof_fees"`
		ProofInterval      int    `toml:"proof_interval"`
	}
	Rollkit struct {
		Port             int `toml:"port"`
		DataShardCount   int `toml:"data_shard_count"`
		ParityShardCount int `toml:"parity_shard_count"`
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

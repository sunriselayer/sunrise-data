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
		AddrPrefix          string `toml:"addr_prefix"`
		PublisherAccount    string `toml:"publisher_account"`
		HomePath            string `toml:"home_path"`
		KeyringBackend      string `toml:"keyring_backend"`
		Fees                string `toml:"fees"`
		CometbftRPC         string `toml:"cometbft_rpc"`
		VoteExtensionPeriod int    `toml:"vote_extension_period"`
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

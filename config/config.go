package config

import (
	toml "github.com/pelletier/go-toml"
)

type Config struct {
	Api struct {
		Port       int    `toml:"port"`
		IpfsApiUrl string `toml:"ipfs_api_url"`
	}
	Chain struct {
		AddrPrefix     string `toml:"addr_prefix"`
		PublisherAccount   string `toml:"publisher_account"`
		HomePath       string `toml:"home_path"`
		KeyringBackend string `toml:"keyring_backend"`
		Fees string `toml:"fees"`
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

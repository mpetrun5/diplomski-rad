package config

import (
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"
)

type Config struct {
	BootstrapPeers []multiaddr.Multiaddr
	Port           uint16
	KeysharePath   string
	Key            string
	Threshold      int
}

func LoadConfig(path string) (*Config, error) {
	rawConfig := Config{}

	viper.SetConfigFile(path)
	viper.SetConfigType("json")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&rawConfig)
	if err != nil {
		return nil, err
	}

	return &rawConfig, nil
}

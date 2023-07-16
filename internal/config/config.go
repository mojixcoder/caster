package config

import (
	"github.com/spf13/viper"
)

// AppConfig holds the entire app configurations.
type AppConfig struct {
	Caster *CasterConfig
	Nodes  []NodeConfig
}

// NodeConfig holds nodes configurations.
type NodeConfig struct {
	Index   int
	Address string
	IsLocal bool `default:"false"`
}

// CasterConfig is the config of Caster.
type CasterConfig struct {
	Capacity uint64 `default:"16384"`
	Port     int    `default:"2376"`
	Debug    bool   `default:"false"`
}

// Load loads the configuration.
func Load() (*AppConfig, error) {
	configPath := viper.GetString("config")

	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

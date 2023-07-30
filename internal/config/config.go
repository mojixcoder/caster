package config

import (
	"github.com/spf13/viper"
)

// AppConfig holds the entire app configurations.
type AppConfig struct {
	Caster *CasterConfig
	Nodes  []NodeConfig
	Tracer TracerConfig
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

// TracerConfig hold tracer configurations.
type TracerConfig struct {
	Name        string  `default:"caster"`
	Fraction    float64 `default:"1"`
	JeagerAgent JaegerAgentConfig
}

// JaegerAgentConfig is the Jaeger agent's config.
type JaegerAgentConfig struct {
	Host string
	Port string `default:"6831"`
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

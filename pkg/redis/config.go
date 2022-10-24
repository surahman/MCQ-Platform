package redis

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/validator"
)

// config is the configuration container for connecting to the Redis cluster
type config struct {
	Authentication struct {
		Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password" validate:"required"`
	} `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
	Connection struct {
		Addrs          []string `json:"addrs,omitempty" yaml:"addrs,omitempty" mapstructure:"addrs" validate:"required,min=1"`
		MaxRedirects   int      `json:"max_redirects,omitempty" yaml:"max_redirects,omitempty" mapstructure:"max_redirects" validate:"required,min=1"`
		MaxRetries     int      `json:"max_retries,omitempty" yaml:"max_retries,omitempty" mapstructure:"max_retries" validate:"required,min=1"`
		PoolSize       int      `json:"pool_size,omitempty" yaml:"pool_size,omitempty" mapstructure:"pool_size" validate:"required,min=1"`
		MinIdleConns   int      `json:"min_idle_conns,omitempty" yaml:"min_idle_conns,omitempty" mapstructure:"min_idle_conns" validate:"required,min=1"`
		ReadOnly       bool     `json:"read_only,omitempty" yaml:"read_only,omitempty" mapstructure:"read_only"`
		RouteByLatency bool     `json:"route_by_latency,omitempty" yaml:"route_by_latency,omitempty" mapstructure:"route_by_latency"`
	} `json:"connection,omitempty" yaml:"connection,omitempty" mapstructure:"connection"`
}

// newConfig creates a blank configuration struct for Redis.
func newConfig() *config {
	return &config{}
}

// Load will attempt to load configurations from a file on a file system and then overwrite values using environment variables.
func (cfg *config) Load(fs afero.Fs) (err error) {
	viper.SetFs(fs)
	viper.SetConfigName(constants.GetRedisFileName())
	viper.SetConfigType("yaml")
	viper.AddConfigPath(constants.GetEtcDir())
	viper.AddConfigPath(constants.GetHomeDir())
	viper.AddConfigPath(constants.GetBaseDir())

	viper.SetEnvPrefix(constants.GetRedisPrefix())
	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	if err = viper.Unmarshal(cfg); err != nil {
		return
	}

	if err = validator.ValidateStruct(cfg); err != nil {
		return
	}

	return
}

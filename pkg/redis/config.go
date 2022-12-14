package redis

import (
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/config_loader"
	"github.com/surahman/mcq-platform/pkg/constants"
)

// config is the configuration container for connecting to the Redis cluster
type config struct {
	Authentication struct {
		Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password" validate:"required"`
	} `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
	Connection struct {
		Addrs           []string `json:"addrs,omitempty" yaml:"addrs,omitempty" mapstructure:"addrs" validate:"required,min=1"`
		MaxConnAttempts int      `json:"max_connection_attempts,omitempty" yaml:"max_connection_attempts,omitempty" mapstructure:"max_connection_attempts" validate:"required,min=1"`
		MaxRedirects    int      `json:"max_redirects,omitempty" yaml:"max_redirects,omitempty" mapstructure:"max_redirects" validate:"required,min=1"`
		MaxRetries      int      `json:"max_retries,omitempty" yaml:"max_retries,omitempty" mapstructure:"max_retries" validate:"required,min=1"`
		PoolSize        int      `json:"pool_size,omitempty" yaml:"pool_size,omitempty" mapstructure:"pool_size" validate:"required,min=1"`
		MinIdleConns    int      `json:"min_idle_conns,omitempty" yaml:"min_idle_conns,omitempty" mapstructure:"min_idle_conns" validate:"required,min=1"`
		ReadOnly        bool     `json:"read_only,omitempty" yaml:"read_only,omitempty" mapstructure:"read_only"`
		RouteByLatency  bool     `json:"route_by_latency,omitempty" yaml:"route_by_latency,omitempty" mapstructure:"route_by_latency"`
	} `json:"connection,omitempty" yaml:"connection,omitempty" mapstructure:"connection"`
	Data struct {
		TTL int64 `json:"ttl,omitempty" yaml:"ttl,omitempty" mapstructure:"ttl" validate:"omitempty,min=60"`
	} `json:"data,omitempty" yaml:"data,omitempty" mapstructure:"data"`
}

// newConfig creates a blank configuration struct for Redis.
func newConfig() *config {
	return &config{}
}

// Load will attempt to load configurations from a file on a file system and then overwrite values using environment variables.
func (cfg *config) Load(fs afero.Fs) (err error) {
	return config_loader.ConfigLoader(fs, cfg, constants.GetRedisFileName(), constants.GetRedisPrefix(), "yaml")
}

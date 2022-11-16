package graphql

import (
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/config_loader"
	"github.com/surahman/mcq-platform/pkg/constants"
)

// config is the configuration container for the HTTP GraphQL endpoint.
type config struct {
	Server struct {
		BasePath       string `json:"base_path,omitempty" yaml:"base_path,omitempty" mapstructure:"base_path" validate:"required"`
		PlaygroundPath string `json:"playground_path,omitempty" yaml:"playground_path,omitempty" mapstructure:"playground_path" validate:"required"`
		QueryPath      string `json:"query_path,omitempty" yaml:"query_path,omitempty" mapstructure:"query_path" validate:"required"`
		PortNumber     int    `json:"port_number,omitempty" yaml:"port_number,omitempty" mapstructure:"port_number" validate:"required,min=1000"`
		ShutdownDelay  int    `json:"shutdown_delay,omitempty" yaml:"shutdown_delay,omitempty" mapstructure:"shutdown_delay" validate:"required,min=0"`
	} `json:"server,omitempty" yaml:"server,omitempty" mapstructure:"server" validate:"required"`
	Authorization struct {
		HeaderKey string `json:"header_key,omitempty" yaml:"header_key,omitempty" mapstructure:"header_key" validate:"required"`
	} `json:"authorization,omitempty" yaml:"authorization,omitempty" mapstructure:"authorization" validate:"required"`
}

// newConfig creates a blank configuration struct for Cassandra.
func newConfig() *config {
	return &config{}
}

// Load will attempt to load configurations from a file on a file system and then overwrite values using environment variables.
func (cfg *config) Load(fs afero.Fs) (err error) {
	return config_loader.ConfigLoader(fs, cfg, constants.GetGraphQLFileName(), constants.GetGraphQLPrefix(), "yaml")
}

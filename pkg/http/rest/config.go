package rest

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/validator"
)

// config is the configuration container for the HTTP REST endpoint.
type config struct {
	Server struct {
		BasePath      string `json:"base_path,omitempty" yaml:"base_path,omitempty" mapstructure:"base_path" validate:"required"`
		SwaggerPath   string `json:"swagger_path,omitempty" yaml:"swagger_path,omitempty" mapstructure:"swagger_path" validate:"required"`
		PortNumber    int    `json:"port_number,omitempty" yaml:"port_number,omitempty" mapstructure:"port_number" validate:"required,min=1000"`
		ShutdownDelay int    `json:"shutdown_delay,omitempty" yaml:"shutdown_delay,omitempty" mapstructure:"shutdown_delay" validate:"required,min=0"`
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
	viper.SetFs(fs)
	viper.SetConfigName(constants.GetHTTPRESTFileName())
	viper.SetConfigType("yaml")
	viper.AddConfigPath(constants.GetEtcDir())
	viper.AddConfigPath(constants.GetHomeDir())
	viper.AddConfigPath(".")

	viper.SetEnvPrefix(constants.GetHTTPRESTPrefix())
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

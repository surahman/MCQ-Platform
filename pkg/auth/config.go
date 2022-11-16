package auth

import (
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/config_loader"
	"github.com/surahman/mcq-platform/pkg/constants"
)

// Config contains all the configurations for authentication.
type config struct {
	JWTConfig struct {
		Key                string `json:"key,omitempty" yaml:"key,omitempty" mapstructure:"key" validate:"required,min=8,max=256"`
		Issuer             string `json:"issuer,omitempty" yaml:"issuer,omitempty" mapstructure:"issuer" validate:"required"`
		ExpirationDuration int64  `json:"expiration_duration,omitempty" yaml:"expiration_duration,omitempty" mapstructure:"expiration_duration" validate:"required,min=60,gtefield=RefreshThreshold"`
		RefreshThreshold   int64  `json:"refresh_threshold,omitempty" yaml:"refresh_threshold,omitempty" mapstructure:"refresh_threshold" validate:"required,min=1,ltefield=ExpirationDuration"`
	} `json:"jwt,omitempty" yaml:"jwt,omitempty" mapstructure:"jwt" validate:"required"`
	General struct {
		BcryptCost   int    `json:"bcrypt_cost,omitempty" yaml:"bcrypt_cost,omitempty" mapstructure:"bcrypt_cost" validate:"required,min=4,max=31"`
		CryptoSecret string `json:"crypto_secret,omitempty" yaml:"crypto_secret,omitempty" mapstructure:"crypto_secret" validate:"required,len=32"`
	} `json:"general,omitempty" yaml:"general,omitempty" mapstructure:"general" validate:"required"`
}

// newConfig creates a blank configuration struct for authorization.
func newConfig() *config {
	return &config{}
}

// Load will attempt to load configurations from a file on a file system and then overwrite values using environment variables.
func (cfg *config) Load(fs afero.Fs) (err error) {
	return config_loader.ConfigLoader(fs, cfg, constants.GetAuthFileName(), constants.GetAuthPrefix(), "yaml")
}

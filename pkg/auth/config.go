package auth

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/validator"
)

// Config contains all the configurations for authentication.
type config struct {
	JWT     *jwtConfig     `json:"jwt,omitempty" yaml:"jwt,omitempty" validate:"required"`
	General *generalConfig `json:"general,omitempty" yaml:"general,omitempty" validate:"required"`
}

// generalConfig is the general configurations for authentication and encryption.
type generalConfig struct {
	BcryptCost int `json:"bcrypt_cost,omitempty" yaml:"bcrypt_cost,omitempty" validate:"required,min=4,max=31"`
}

// jwtConfig are the configurations for the JSON Web Tokens.
type jwtConfig struct {
	Key                string `json:"key,omitempty" yaml:"key,omitempty" validate:"required,min=8,max=256"`
	ExpirationDuration int    `json:"expiration_duration,omitempty" yaml:"expiration_duration,omitempty" validate:"required,min=10"`
}

// newConfig creates a blank configuration struct for authorization.
func newConfig() *config {
	return &config{}
}

// Load will attempt to load configurations from a file on a file system and then overwrite values using environment variables.
func (cfg *config) Load(fs afero.Fs) (err error) {
	viper.SetFs(fs)
	viper.SetConfigName(constants.GetAuthFileName())
	viper.SetConfigType("yaml")
	viper.AddConfigPath(constants.GetEtcDir())
	viper.AddConfigPath(constants.GetHomeDir())
	viper.AddConfigPath(".")

	viper.SetEnvPrefix(constants.GetAuthPrefix())
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

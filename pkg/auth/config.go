package auth

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/validator"
)

// Config contains all the configurations for authentication.
type config struct {
	JWTConfig struct {
		Key                string `json:"key,omitempty" yaml:"key,omitempty" mapstructure:"key" validate:"required,min=8,max=256"`
		Issuer             string `json:"issuer,omitempty" yaml:"issuer,omitempty" mapstructure:"issuer" validate:"required"`
		ExpirationDuration int64  `json:"expiration_duration,omitempty" yaml:"expiration_duration,omitempty" mapstructure:"expiration_duration" validate:"required,min=10"`
	} `json:"jwt,omitempty" yaml:"jwt,omitempty" mapstructure:"jwt" validate:"required"`
	General struct {
		BcryptCost int `json:"bcrypt_cost,omitempty" yaml:"bcrypt_cost,omitempty" mapstructure:"bcrypt_cost" validate:"required,min=4,max=31"`
	} `json:"general,omitempty" yaml:"general,omitempty" mapstructure:"general" validate:"required"`
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
	viper.AddConfigPath(constants.GetBaseDir())

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

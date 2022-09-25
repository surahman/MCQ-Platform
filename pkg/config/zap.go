package config

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/surahman/cassandra-tutorial/pkg/validator"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapConfig struct {
	BuiltinConfig        string `json:"builtin_config,omitempty" yaml:"builtin_config,omitempty" validator:"required,oneof='Production' 'Development'"`
	BuiltinEncoderConfig string `json:"builtin_encoder_config,omitempty" yaml:"builtin_encoder_config,omitempty" validator:"required,oneof='Production' 'Development'"`
	GeneralConfig        struct {
		zap.Config
	} `json:"general_config,omitempty" yaml:"general_config,omitempty" validator:"required_without=BuiltinConfig"`
	EncoderConfig struct {
		zapcore.EncoderConfig
	} `json:"encoder_config,omitempty" yaml:"encoder_config,omitempty" validator:"required_without=EncoderConfig"`
}

// Load will attempt to load configurations from a file on a file system.
func (cfg *ZapConfig) Load(fs afero.Fs) (err error) {
	viper.SetFs(fs)
	viper.SetConfigName(GetLoggerFileName())
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configEtcDir)
	viper.AddConfigPath(configHomeDir)
	viper.AddConfigPath(".")

	viper.SetEnvPrefix(loggerPrefix)
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

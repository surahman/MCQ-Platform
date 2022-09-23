package config

import (
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapConfig struct {
	BuiltinConfig        string `json:"builtin_config,omitempty" yaml:"builtin_config,omitempty" validator:"oneof='Production' 'Development'"`
	BuiltinEncoderConfig string `json:"builtin_encoder_config,omitempty" yaml:"builtin_encoder_config,omitempty" validator:"oneof='Production' 'Development'"`
	GeneralConfig        struct {
		zap.Config
	} `json:"general_config,omitempty" yaml:"general_config,omitempty"`
	EncoderConfig struct {
		zapcore.EncoderConfig
	} `json:"encoder_config,omitempty" yaml:"encoder_config,omitempty"`
}

// Load will attempt to load configurations from a file on a file system.
func (c *ZapConfig) Load(afero.Fs) (err error) {
	return
}

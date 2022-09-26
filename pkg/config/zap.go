package config

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/surahman/mcq-platform/pkg/validator"
)

type ZapConfig struct {
	BuiltinConfig        string `json:"builtin_config,omitempty" yaml:"builtin_config,omitempty" mapstructure:"builtin_config" validate:"oneof='Production' 'Development'"`
	BuiltinEncoderConfig string `json:"builtin_encoder_config,omitempty" yaml:"builtin_encoder_config,omitempty" mapstructure:"builtin_encoder_config" validate:"oneof='Production' 'Development''"`
	GeneralConfig        struct {
		Development       *bool     `json:"development,omitempty" yaml:"development,omitempty" mapstructure:"development"`
		DisableCaller     *bool     `json:"disableCaller,omitempty" yaml:"disableCaller,omitempty" mapstructure:"disableCaller"`
		DisableStacktrace *bool     `json:"disableStacktrace,omitempty" yaml:"disableStacktrace,omitempty" mapstructure:"disableStacktrace"`
		Encoding          *string   `json:"encoding,omitempty" yaml:"encoding,omitempty" mapstructure:"encoding"`
		OutputPaths       *[]string `json:"outputPaths,omitempty" yaml:"outputPaths,omitempty" mapstructure:"outputPaths"`
		ErrorOutputPaths  *[]string `json:"errorOutputPaths,omitempty" yaml:"errorOutputPaths,omitempty" mapstructure:"errorOutputPaths"`
	} `json:"general_config,omitempty" yaml:"general_config,omitempty" mapstructure:"general_config"`
	EncoderConfig struct {
		MessageKey       *string `json:"messageKey,omitempty" yaml:"messageKey,omitempty" mapstructure:"messageKey"`
		LevelKey         *string `json:"levelKey,omitempty" yaml:"levelKey,omitempty" mapstructure:"levelKey"`
		TimeKey          *string `json:"timeKey,omitempty" yaml:"timeKey,omitempty" mapstructure:"timeKey"`
		NameKey          *string `json:"nameKey,omitempty" yaml:"nameKey,omitempty" mapstructure:"nameKey"`
		CallerKey        *string `json:"callerKey,omitempty" yaml:"callerKey,omitempty" mapstructure:"callerKey"`
		FunctionKey      *string `json:"functionKey,omitempty" yaml:"functionKey,omitempty" mapstructure:"functionKey"`
		StacktraceKey    *string `json:"stacktraceKey,omitempty" yaml:"stacktraceKey,omitempty" mapstructure:"stacktraceKey"`
		SkipLineEnding   *bool   `json:"skipLineEnding,omitempty" yaml:"skipLineEnding,omitempty" mapstructure:"skipLineEnding"`
		LineEnding       *string `json:"lineEnding,omitempty" yaml:"lineEnding,omitempty" mapstructure:"lineEnding"`
		ConsoleSeparator *string `json:"consoleSeparator,omitempty" yaml:"consoleSeparator,omitempty" mapstructure:"consoleSeparator"`
	} `json:"encoder_config,omitempty" yaml:"encoder_config,omitempty" mapstructure:"encoder_config"`
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

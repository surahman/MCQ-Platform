package logger

import (
	"log"

	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

var zapLogger *zap.Logger

func Init(fs *afero.Fs) {
	var err error
	var userConfig *config.ZapConfig
	var baseConfig zap.Config
	var encConfig zapcore.EncoderConfig

	{
		cfg, err := config.Factory(config.Logger)
		if err != nil {
			log.Fatalf("failed to load logger configuration file from disk: %v", err)
		}
		userConfig = cfg.(*config.ZapConfig)
		if err = cfg.Load(*fs); err != nil {
			log.Fatalf("failed to load logger configuration file from disk: %v", err)
		}
	}

	// Base logger configuration.
	switch userConfig.BuiltinConfig {
	case "Development":
		baseConfig = zap.NewDevelopmentConfig()
		break
	case "Production":
		baseConfig = zap.NewProductionConfig()
		break
	default:
		log.Fatal("could not select the base config type")
	}

	// Encoder logger configuration.
	switch userConfig.BuiltinEncoderConfig {
	case "Development":
		encConfig = zap.NewDevelopmentEncoderConfig()
		break
	case "Production":
		encConfig = zap.NewProductionEncoderConfig()
		break
	default:
		log.Fatal("could not select the base encoder config type")
	}

	if err = mergeConfig[*zap.Config, *config.ZapGeneralConfig](&baseConfig, userConfig.GeneralConfig); err != nil {
		log.Fatalf("failed to merge base configurations and user provided configurations for logger")
	}
	if err = mergeConfig[*zapcore.EncoderConfig, *config.ZapEncoderConfig](&encConfig, userConfig.EncoderConfig); err != nil {
		log.Fatalf("failed to merge base encoder configurations and user provided encoder configurations for logger")
	}

	baseConfig.EncoderConfig = encConfig
	if zapLogger, err = baseConfig.Build(zap.AddCallerSkip(1)); err != nil {
		log.Fatalf("failure configuring logger: %v", err)
	}
}

// Info logs messages at the info level.
func Info(message string, fields ...zap.Field) {
	zapLogger.Info(message, fields...)
}

// Debug logs messages at the debug level.
func Debug(message string, fields ...zap.Field) {
	zapLogger.Debug(message, fields...)
}

// Error logs messages at the error level.
func Error(message string, fields ...zap.Field) {
	zapLogger.Error(message, fields...)
}

// Fatal logs messages at the fatal level.
func Fatal(message string, fields ...zap.Field) {
	zapLogger.Fatal(message, fields...)
}

// mergeConfig will merge the configuration files by marshalling and unmarshalling.
func mergeConfig[DST *zap.Config | *zapcore.EncoderConfig, SRC *config.ZapGeneralConfig | *config.ZapEncoderConfig](dst DST, src SRC) (err error) {
	var yamlToConv []byte
	if yamlToConv, err = yaml.Marshal(src); err != nil {
		return
	}
	if err = yaml.Unmarshal(yamlToConv, dst); err != nil {
		return
	}
	return
}

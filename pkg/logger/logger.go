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

func Init(fs *afero.Fs) (err error) {
	var baseConfig zap.Config
	var encConfig zapcore.EncoderConfig

	userConfig := config.NewLoggerConfig()
	if err = userConfig.Load(*fs); err != nil {
		log.Printf("failed to load logger configuration file from disk: %v\n", err)
		return
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
		log.Println("could not select the base config type")
		return
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
		log.Println("could not select the base encoder config type")
		return
	}

	if err = mergeConfig[*zap.Config, *config.ZapGeneralConfig](&baseConfig, userConfig.GeneralConfig); err != nil {
		log.Printf("failed to merge base configurations and user provided configurations for logger: %v\n", err)
		return
	}
	if err = mergeConfig[*zapcore.EncoderConfig, *config.ZapEncoderConfig](&encConfig, userConfig.EncoderConfig); err != nil {
		log.Printf("failed to merge base encoder configurations and user provided encoder configurations for logger: %v\n", err)
		return
	}

	baseConfig.EncoderConfig = encConfig
	if zapLogger, err = baseConfig.Build(zap.AddCallerSkip(1)); err != nil {
		log.Printf("failure configuring logger: %v\n", err)
		return
	}
	return
}

// Info logs messages at the info level.
func Info(message string, fields ...zap.Field) {
	zapLogger.Info(message, fields...)
}

// Debug logs messages at the debug level.
func Debug(message string, fields ...zap.Field) {
	zapLogger.Debug(message, fields...)
}

// Warn logs messages at the warn level.
func Warn(message string, fields ...zap.Field) {
	zapLogger.Warn(message, fields...)
}

// Error logs messages at the error level.
func Error(message string, fields ...zap.Field) {
	zapLogger.Error(message, fields...)
}

// Panic logs messages at the panic level and then panics at the call site.
func Panic(message string, fields ...zap.Field) {
	zapLogger.Panic(message, fields...)
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

package logger

import (
	"errors"
	"log"
	"strings"

	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

// Logger is the Zap logger object.
type Logger struct {
	zapLogger *zap.Logger
}

// NewLogger will create a new uninitialized logger.
func NewLogger() *Logger {
	return &Logger{}
}

// Init will initialize the logger with configurations and start it.
func (l *Logger) Init(fs *afero.Fs) (err error) {
	if l.zapLogger != nil {
		return errors.New("logger is already initialized")
	}
	var baseConfig zap.Config
	var encConfig zapcore.EncoderConfig

	userConfig := config.NewLoggerConfig()
	if err = userConfig.Load(*fs); err != nil {
		log.Printf("failed to load logger configuration file from disk: %v\n", err)
		return
	}

	// Base logger configuration.
	switch strings.ToLower(userConfig.BuiltinConfig) {
	case "development":
		baseConfig = zap.NewDevelopmentConfig()
		break
	case "production":
		baseConfig = zap.NewProductionConfig()
		break
	default:
		log.Println("could not select the base config type")
		return
	}

	// Encoder logger configuration.
	switch strings.ToLower(userConfig.BuiltinEncoderConfig) {
	case "development":
		encConfig = zap.NewDevelopmentEncoderConfig()
		break
	case "production":
		encConfig = zap.NewProductionEncoderConfig()
		break
	default:
		log.Println("could not select the base encoder config type")
		return
	}

	// Merge configurations.
	if err = mergeConfig[*zap.Config, *config.ZapGeneralConfig](&baseConfig, userConfig.GeneralConfig); err != nil {
		log.Printf("failed to merge base configurations and user provided configurations for logger: %v\n", err)
		return
	}
	if err = mergeConfig[*zapcore.EncoderConfig, *config.ZapEncoderConfig](&encConfig, userConfig.EncoderConfig); err != nil {
		log.Printf("failed to merge base encoder configurations and user provided encoder configurations for logger: %v\n", err)
		return
	}

	// Init and create logger.
	baseConfig.EncoderConfig = encConfig
	if l.zapLogger, err = baseConfig.Build(zap.AddCallerSkip(1)); err != nil {
		log.Printf("failure configuring logger: %v\n", err)
		return
	}
	return
}

// Info logs messages at the info level.
func (l *Logger) Info(message string, fields ...zap.Field) {
	l.zapLogger.Info(message, fields...)
}

// Debug logs messages at the debug level.
func (l *Logger) Debug(message string, fields ...zap.Field) {
	l.zapLogger.Debug(message, fields...)
}

// Warn logs messages at the warn level.
func (l *Logger) Warn(message string, fields ...zap.Field) {
	l.zapLogger.Warn(message, fields...)
}

// Error logs messages at the error level.
func (l *Logger) Error(message string, fields ...zap.Field) {
	l.zapLogger.Error(message, fields...)
}

// Panic logs messages at the panic level and then panics at the call site.
func (l *Logger) Panic(message string, fields ...zap.Field) {
	l.zapLogger.Panic(message, fields...)
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

/*** Unexported methods and members to support testing. ***/

// setTestLogger is an unexported utility method that set a logger base for testing.
func (l *Logger) setTestLogger(testLogger *zap.Logger) {
	l.zapLogger = testLogger
}

// NewTestLogger will create a new development logger to be used in test suites.
func NewTestLogger() (logger *Logger, err error) {
	baseConfig := zap.NewDevelopmentConfig()
	baseConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	var zapLogger *zap.Logger
	if zapLogger, err = baseConfig.Build(zap.AddCallerSkip(1)); err != nil {
		log.Printf("failure configuring logger: %v\n", err)
		return nil, err
	}
	return &Logger{zapLogger: zapLogger}, err
}

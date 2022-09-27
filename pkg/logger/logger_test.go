package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestMergeConfig_General(t *testing.T) {
	userGenCfg := config.ZapGeneralConfig{
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Encoding:          "reset",
		OutputPaths:       []string{"stdout", "/etc/appname.log"},
		ErrorOutputPaths:  []string{"stderr", "/etc/appname_err.log"},
	}
	zapCfg := zap.NewDevelopmentConfig()
	require.NoError(t, mergeConfig[*zap.Config, *config.ZapGeneralConfig](&zapCfg, &userGenCfg), "Failed to merge config files.")
	require.Equalf(t, userGenCfg.Development, zapCfg.Development, "Development value expected %v, actual %v", userGenCfg.Development, zapCfg.Development)
	require.Equalf(t, userGenCfg.DisableCaller, zapCfg.DisableCaller, "DisableCaller value expected %v, actual %v", userGenCfg.DisableCaller, zapCfg.DisableCaller)
	require.Equalf(t, userGenCfg.DisableStacktrace, zapCfg.DisableStacktrace, "DisableStacktrace value expected %v, actual %v", userGenCfg.DisableStacktrace, zapCfg.DisableStacktrace)
	require.Equalf(t, userGenCfg.Encoding, zapCfg.Encoding, "Encoding value expected %v, actual %v", userGenCfg.Encoding, zapCfg.Encoding)
	require.Equalf(t, userGenCfg.OutputPaths, zapCfg.OutputPaths, "OutputPaths value expected %v, actual %v", userGenCfg.OutputPaths, zapCfg.OutputPaths)
	require.Equalf(t, userGenCfg.ErrorOutputPaths, zapCfg.ErrorOutputPaths, "ErrorOutputPaths value expected %v, actual %v", userGenCfg.ErrorOutputPaths, zapCfg.ErrorOutputPaths)
}

func TestMergeConfig_Encoder(t *testing.T) {
	userEncCfg := config.ZapEncoderConfig{
		MessageKey:       "message key",
		LevelKey:         "level key",
		TimeKey:          "time key",
		NameKey:          "name key",
		CallerKey:        "caller key",
		FunctionKey:      "function key",
		StacktraceKey:    "stack trace key",
		SkipLineEnding:   true,
		LineEnding:       "line ending",
		ConsoleSeparator: "console separator",
	}
	zapCfg := zap.NewDevelopmentEncoderConfig()
	require.NoError(t, mergeConfig[*zapcore.EncoderConfig, *config.ZapEncoderConfig](&zapCfg, &userEncCfg), "Failed to merge config files.")
	require.Equalf(t, userEncCfg.MessageKey, zapCfg.MessageKey, "MessageKey value expected %v, actual %v", userEncCfg.MessageKey, zapCfg.MessageKey)
	require.Equalf(t, userEncCfg.LevelKey, zapCfg.LevelKey, "LevelKey value expected %v, actual %v", userEncCfg.LevelKey, zapCfg.LevelKey)
	require.Equalf(t, userEncCfg.TimeKey, zapCfg.TimeKey, "TimeKey value expected %v, actual %v", userEncCfg.TimeKey, zapCfg.TimeKey)
	require.Equalf(t, userEncCfg.NameKey, zapCfg.NameKey, "NameKey value expected %v, actual %v", userEncCfg.NameKey, zapCfg.NameKey)
	require.Equalf(t, userEncCfg.CallerKey, zapCfg.CallerKey, "CallerKey value expected %v, actual %v", userEncCfg.CallerKey, zapCfg.CallerKey)
	require.Equalf(t, userEncCfg.FunctionKey, zapCfg.FunctionKey, "FunctionKey value expected %v, actual %v", userEncCfg.FunctionKey, zapCfg.FunctionKey)
	require.Equalf(t, userEncCfg.StacktraceKey, zapCfg.StacktraceKey, "StacktraceKey value expected %v, actual %v", userEncCfg.StacktraceKey, zapCfg.StacktraceKey)
	require.Equalf(t, userEncCfg.SkipLineEnding, zapCfg.SkipLineEnding, "SkipLineEnding value expected %v, actual %v", userEncCfg.SkipLineEnding, zapCfg.SkipLineEnding)
	require.Equalf(t, userEncCfg.LineEnding, zapCfg.LineEnding, "LineEnding value expected %v, actual %v", userEncCfg.LineEnding, zapCfg.LineEnding)
	require.Equalf(t, userEncCfg.ConsoleSeparator, zapCfg.ConsoleSeparator, "ConsoleSeparator value expected %v, actual %v", userEncCfg.ConsoleSeparator, zapCfg.ConsoleSeparator)
}

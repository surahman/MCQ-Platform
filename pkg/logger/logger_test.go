package logger

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

var loggerConfigTestData = config.LoggerConfigTestData()

func TestNewLogger(t *testing.T) {
	require.Equal(t, reflect.TypeOf(NewLogger()), reflect.TypeOf(&Logger{}), "creates new logger successfully")
}

func TestNewTestLogger(t *testing.T) {
	testLogger, err := NewTestLogger()
	require.NoError(t, err, "failed to create new logger for use in test suites")
	require.Equal(t, reflect.TypeOf(testLogger), reflect.TypeOf(&zap.Logger{}), "creates new logger successfully")
}

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

func TestInit(t *testing.T) {
	fullFilePath := config.GetEtcDir() + config.GetLoggerFileName()

	testCases := []struct {
		name      string
		input     string
		expectErr require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"invalid - empty",
			loggerConfigTestData["empty"],
			require.Error,
		}, {
			"valid - development",
			loggerConfigTestData["valid_devel"],
			require.NoError,
		}, {
			"valid - production",
			loggerConfigTestData["valid_prod"],
			require.NoError,
		}, {
			"valid full - development",
			loggerConfigTestData["valid_config"],
			require.NoError,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Init mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(config.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, fullFilePath, []byte(testCase.input), 0644), "Failed to write in memory file")

			logger := NewLogger()
			testCase.expectErr(t, logger.Init(&fs), "Error condition failed when trying to initialize logger")
		})
	}
}

func TestTestLogger(t *testing.T) {
	ts := newTestLogSpy(t)
	defer ts.AssertPassed()

	logger := NewLogger()
	logger.setTestLogger(zaptest.NewLogger(ts))

	logger.Info("info message received")
	logger.Debug("debug message received")
	logger.Warn("warn message received")
	logger.Error("error message received", zap.Error(errors.New("ow no! woe is me!")))

	assert.Panics(t, func() {
		logger.Panic("panic message received")
	}, "Panic should panic")

	ts.AssertMessages(
		"INFO	info message received",
		"DEBUG	debug message received",
		"WARN	warn message received",
		`ERROR	error message received	{"error": "ow no! woe is me!"}`,
		"PANIC	panic message received",
	)
}

func TestTestLoggerSupportsLevels(t *testing.T) {
	ts := newTestLogSpy(t)
	defer ts.AssertPassed()

	logger := NewLogger()
	logger.setTestLogger(zaptest.NewLogger(ts, zaptest.Level(zap.WarnLevel)))

	logger.Info("info message received")
	logger.Debug("debug message received")
	logger.Warn("warn message received")
	logger.Error("error message received", zap.Error(errors.New("ow no! woe is me!")))

	assert.Panics(t, func() {
		logger.Panic("panic message received")
	}, "Panic should panic")

	ts.AssertMessages(
		"WARN	warn message received",
		`ERROR	error message received	{"error": "ow no! woe is me!"}`,
		"PANIC	panic message received",
	)
}

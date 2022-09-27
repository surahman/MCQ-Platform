package config

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var loggerConfigTestData = LoggerConfigTestData()

func TestZapConfig_Load(t *testing.T) {
	envCfgKey := fmt.Sprintf("%s_BUILTIN_CONFIG", loggerPrefix)
	envEncKey := fmt.Sprintf("%s_BUILTIN_ENCODER_CONFIG", loggerPrefix)

	testCases := []struct {
		name      string
		fullPath  string
		input     string
		cfgKey    string
		encKey    string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"invalid - empty",
			configEtcDir,
			loggerConfigTestData["empty"],
			"Production",
			"Production",
			require.Error,
			require.Nil,
		}, {
			"invalid - builtin",
			configEtcDir,
			loggerConfigTestData["invalid_builtin"],
			xid.New().String(),
			xid.New().String(),
			require.Error,
			require.Nil,
		}, {
			"valid - development",
			configEtcDir,
			loggerConfigTestData["valid_devel"],
			"Production",
			"Production",
			require.NoError,
			require.Nil,
		}, {
			"valid - production",
			configEtcDir,
			loggerConfigTestData["valid_prod"],
			"Development",
			"Development",
			require.NoError,
			require.Nil,
		}, {
			"valid - full config",
			configEtcDir,
			loggerConfigTestData["valid_config"],
			"Production",
			"Production",
			require.NoError,
			require.NotNil,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(testCase.fullPath, 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, testCase.fullPath+loggerConfigFileName, []byte(testCase.input), 0644), "Failed to write in memory file")

			// Load from mock filesystem.
			actual := &ZapConfig{}
			err := actual.Load(fs)
			testCase.expectErr(t, err)

			if err != nil {
				return
			}

			// Load expected struct.
			expected := &ZapConfig{}
			require.NoError(t, yaml.Unmarshal([]byte(testCase.input), expected), "failed to unmarshal expected config")
			require.True(t, reflect.DeepEqual(expected, actual))

			// Test configuring of environment variable.
			require.NoErrorf(t, os.Setenv(envCfgKey, testCase.cfgKey), "Failed to set environment variable for config: %v", err)
			require.NoErrorf(t, os.Setenv(envEncKey, testCase.encKey), "Failed to set environment variable for encoder: %v", err)
			require.NoErrorf(t, actual.Load(fs), "Failed to load config file: %v", err)
			require.NoErrorf(t, os.Unsetenv(envCfgKey), "Failed to unset environment variable set for config")
			require.NoErrorf(t, os.Unsetenv(envEncKey), "Failed to unset environment variable set for encoder")

			require.Equalf(t, testCase.cfgKey, actual.BuiltinConfig, "Failed to load environment variable into config")
			require.Equalf(t, testCase.encKey, actual.BuiltinEncoderConfig, "Failed to load environment variable into encoder")

			testCase.expectNil(t, actual.GeneralConfig, "Check for nil general config failed")
		})
	}
}

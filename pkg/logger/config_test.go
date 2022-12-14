package logger

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/constants"
	"gopkg.in/yaml.v3"
)

func TestZapConfig_Load(t *testing.T) {
	envCfgKey := fmt.Sprintf("%s_BUILTIN_CONFIG", constants.GetLoggerPrefix())
	envEncKey := fmt.Sprintf("%s_BUILTIN_ENCODER_CONFIG", constants.GetLoggerPrefix())

	testCases := []struct {
		name      string
		input     string
		cfgKey    string
		encKey    string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"invalid - empty",
			loggerConfigTestData["empty"],
			"Production",
			"Production",
			require.Error,
			require.Nil,
		}, {
			"invalid - builtin",
			loggerConfigTestData["invalid_builtin"],
			xid.New().String(),
			xid.New().String(),
			require.Error,
			require.Nil,
		}, {
			"valid - development",
			loggerConfigTestData["valid_devel"],
			"Production",
			"Production",
			require.NoError,
			require.Nil,
		}, {
			"valid - production",
			loggerConfigTestData["valid_prod"],
			"Development",
			"Development",
			require.NoError,
			require.Nil,
		}, {
			"valid - full constants",
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
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetLoggerFileName(), []byte(testCase.input), 0644), "Failed to write in memory file")

			// Load from mock filesystem.
			actual := &config{}
			err := actual.Load(fs)
			testCase.expectErr(t, err)

			if err != nil {
				return
			}

			// Load expected struct.
			expected := &config{}
			require.NoError(t, yaml.Unmarshal([]byte(testCase.input), expected), "failed to unmarshal expected constants")
			require.True(t, reflect.DeepEqual(expected, actual))

			// Test configuring of environment variable.
			t.Setenv(envCfgKey, testCase.cfgKey)
			t.Setenv(envEncKey, testCase.encKey)
			require.NoErrorf(t, actual.Load(fs), "Failed to load constants file: %v", err)

			require.Equalf(t, testCase.cfgKey, actual.BuiltinConfig, "Failed to load environment variable into constants")
			require.Equalf(t, testCase.encKey, actual.BuiltinEncoderConfig, "Failed to load environment variable into encoder")

			testCase.expectNil(t, actual.GeneralConfig, "Check for nil general constants failed")
		})
	}
}

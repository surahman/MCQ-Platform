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

func TestZapConfig_Load(t *testing.T) {

	testData := LoggerConfigTestData()
	keyspaceKey := fmt.Sprintf("%s_GENERAL_CONFIG.ENCODING", loggerPrefix)

	testCases := []struct {
		name      string
		fullPath  string
		input     string
		envValue  string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"invalid - empty",
			configEtcDir,
			testData["empty"],
			xid.New().String(),
			require.Error,
			require.Nil,
		},
		{
			"invalid - builtin",
			configEtcDir,
			testData["invalid_builtin"],
			xid.New().String(),
			require.Error,
			require.Nil,
		},
		{
			"valid - development",
			configEtcDir,
			testData["valid_devel"],
			xid.New().String(),
			require.NoError,
			require.Nil,
		},
		{
			"valid - production",
			configEtcDir,
			testData["valid_prod"],
			xid.New().String(),
			require.NoError,
			require.Nil,
		},
		{
			"valid - full config",
			configEtcDir,
			testData["valid_config"],
			xid.New().String(),
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
			require.NoErrorf(t, os.Setenv(keyspaceKey, testCase.envValue), "Failed to set environment variable: %v", err)
			require.NoErrorf(t, actual.Load(fs), "Failed to load config file: %v", err)
			require.NoErrorf(t, os.Unsetenv(keyspaceKey), "Failed to unset environment variable set for test")

			testCase.expectNil(t, actual.GeneralConfig.Encoding, "Check for nil value after environment variable loading")

			if actual.GeneralConfig.Encoding != nil {
				require.Equalf(t, testCase.envValue, *actual.GeneralConfig.Encoding, "Failed to load environment variable into configs")
			}
		})
	}

}

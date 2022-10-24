package redis

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

func TestRedisConfigs_Load(t *testing.T) {
	keyspaceKey := fmt.Sprintf("%s_AUTHENTICATION.PASSWORD", constants.GetRedisPrefix())

	testCases := []struct {
		name      string
		input     string
		envValue  string
		expectErr require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"empty - etc dir",
			redisConfigTestData["empty"],
			xid.New().String(),
			require.Error,
		}, {
			"valid - etc dir",
			redisConfigTestData["valid"],
			xid.New().String(),
			require.NoError,
		}, {
			"no password - etc dir",
			redisConfigTestData["password_empty"],
			xid.New().String(),
			require.Error,
		}, {
			"no addrs - etc dir",
			redisConfigTestData["no_addrs"],
			xid.New().String(),
			require.Error,
		}, {
			"invalid max redirects - etc dir",
			redisConfigTestData["invalid_max_redirects"],
			xid.New().String(),
			require.Error,
		}, {
			"invalid max retries - etc dir",
			redisConfigTestData["invalid_max_retries"],
			xid.New().String(),
			require.Error,
		}, {
			"invalid pool size - etc dir",
			redisConfigTestData["invalid_pool_size"],
			xid.New().String(),
			require.Error,
		}, {
			"invalid min idle conns - etc dir",
			redisConfigTestData["invalid_min_idle_conns"],
			xid.New().String(),
			require.Error,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetRedisFileName(), []byte(testCase.input), 0644), "Failed to write in memory file")

			// Load from mock filesystem.
			actual := &config{}
			err := actual.Load(fs)
			testCase.expectErr(t, err)

			if err != nil {
				return
			}

			// Load expected struct.
			expected := &config{}
			require.NoError(t, yaml.Unmarshal([]byte(testCase.input), expected), "failed to unmarshal expected configurations")
			require.True(t, reflect.DeepEqual(expected, actual))

			// Test configuring of environment variable.
			t.Setenv(keyspaceKey, testCase.envValue)
			require.NoErrorf(t, actual.Load(fs), "Failed to load configurations file: %v", err)
			require.Equalf(t, testCase.envValue, actual.Authentication.Password, "Failed to load environment variable into configs")
		})
	}
}

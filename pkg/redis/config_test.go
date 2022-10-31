package redis

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/validator"
	"gopkg.in/yaml.v3"
)

func TestRedisConfigs_Load(t *testing.T) {
	keyspaceKey := fmt.Sprintf("%s_AUTHENTICATION.PASSWORD", constants.GetRedisPrefix())

	testCases := []struct {
		name         string
		input        string
		envValue     string
		expectErrCnt int
		expectErr    require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name:         "empty - etc dir",
			input:        redisConfigTestData["empty"],
			envValue:     xid.New().String(),
			expectErrCnt: 6,
			expectErr:    require.Error,
		}, {
			name:         "valid - etc dir",
			input:        redisConfigTestData["valid"],
			envValue:     xid.New().String(),
			expectErrCnt: 0,
			expectErr:    require.NoError,
		}, {
			name:         "no password - etc dir",
			input:        redisConfigTestData["password_empty"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "no addrs - etc dir",
			input:        redisConfigTestData["no_addrs"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "invalid max redirects - etc dir",
			input:        redisConfigTestData["invalid_max_redirects"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "invalid max retries - etc dir",
			input:        redisConfigTestData["invalid_max_retries"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "invalid pool size - etc dir",
			input:        redisConfigTestData["invalid_pool_size"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "invalid min idle conns - etc dir",
			input:        redisConfigTestData["invalid_min_idle_conns"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		}, {
			name:         "invalid min TTL - etc dir",
			input:        redisConfigTestData["invalid_min_ttl"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
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
				errorList := err.(*validator.ErrorValidation).Errors
				require.Equalf(t, testCase.expectErrCnt, len(errorList), "expected error count does not match: %v", errorList)
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

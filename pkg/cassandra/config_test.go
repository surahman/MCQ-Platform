package cassandra

import (
	"fmt"
	"github.com/surahman/mcq-platform/pkg/validator"
	"reflect"
	"testing"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/constants"
	"gopkg.in/yaml.v3"
)

func TestCassandraConfigs_Load(t *testing.T) {
	keyspaceKey := fmt.Sprintf("%s_KEYSPACE.REPLICATION_CLASS", constants.GetCassandraPrefix())

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
			input:        cassandraConfigTestData["empty"],
			envValue:     xid.New().String(),
			expectErrCnt: 10,
			expectErr:    require.Error,
		},
		{
			name:         "valid - etc dir",
			input:        cassandraConfigTestData["valid"],
			envValue:     xid.New().String(),
			expectErrCnt: 0,
			expectErr:    require.NoError,
		},
		{
			name:         "no password - etc dir",
			input:        cassandraConfigTestData["password_empty"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		},
		{
			name:         "no username - etc dir",
			input:        cassandraConfigTestData["username_empty"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		},
		{
			name:         "no keyspace - etc dir",
			input:        cassandraConfigTestData["keyspace_empty"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		},
		{
			name:         "no consistency - etc dir",
			input:        cassandraConfigTestData["consistency_missing"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		},
		{
			name:         "no ip - etc dir",
			input:        cassandraConfigTestData["ip_empty"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		},
		{
			name:         "timeout zero - etc dir",
			input:        cassandraConfigTestData["timeout_zero"],
			envValue:     xid.New().String(),
			expectErrCnt: 1,
			expectErr:    require.Error,
		},
		{
			name:         "invalid max connection attempts - etc dir",
			input:        cassandraConfigTestData["invalid_min_max_conn_attempts"],
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
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetCassandraFileName(), []byte(testCase.input), 0644), "Failed to write in memory file")

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
			require.NoError(t, yaml.Unmarshal([]byte(testCase.input), expected), "failed to unmarshal expected constants")
			require.True(t, reflect.DeepEqual(expected, actual))

			// Test configuring of environment variable.
			t.Setenv(keyspaceKey, testCase.envValue)
			require.NoErrorf(t, actual.Load(fs), "Failed to load constants file: %v", err)
			require.Equalf(t, testCase.envValue, actual.Keyspace.ReplicationClass, "Failed to load environment variable into configs")
		})
	}
}

package cassandra

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/config"
	"gopkg.in/yaml.v3"
)

func TestCassandraConfigs_Load(t *testing.T) {
	keyspaceKey := fmt.Sprintf("%s_KEYSPACE.REPLICATION_CLASS", config.GetCassandraPrefix())

	testCases := []struct {
		name      string
		input     string
		envValue  string
		expectErr require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"empty - etc dir",
			cassandraConfigTestData["empty"],
			xid.New().String(),
			require.Error,
		},
		{
			"valid - etc dir",
			cassandraConfigTestData["valid"],
			xid.New().String(),
			require.NoError,
		},
		{
			"no password - etc dir",
			cassandraConfigTestData["password_empty"],
			xid.New().String(),
			require.Error,
		},
		{
			"no username - etc dir",
			cassandraConfigTestData["username_empty"],
			xid.New().String(),
			require.Error,
		},
		{
			"no keyspace - etc dir",
			cassandraConfigTestData["keyspace_empty"],
			xid.New().String(),
			require.Error,
		},
		{
			"no consistency - etc dir",
			cassandraConfigTestData["consistency_missing"],
			xid.New().String(),
			require.Error,
		},
		{
			"no ip - etc dir",
			cassandraConfigTestData["ip_empty"],
			xid.New().String(),
			require.Error,
		},
		{
			"timeout zero - etc dir",
			cassandraConfigTestData["timeout_zero"],
			xid.New().String(),
			require.Error,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(config.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, config.GetEtcDir()+config.GetCassandraFileName(), []byte(testCase.input), 0644), "Failed to write in memory file")

			// Load from mock filesystem.
			actual := &Config{}
			err := actual.Load(fs)
			testCase.expectErr(t, err)

			if err != nil {
				return
			}

			// Load expected struct.
			expected := &Config{}
			require.NoError(t, yaml.Unmarshal([]byte(testCase.input), expected), "failed to unmarshal expected config")
			require.True(t, reflect.DeepEqual(expected, actual))

			// Test configuring of environment variable.
			require.NoErrorf(t, os.Setenv(keyspaceKey, testCase.envValue), "Failed to set environment variable: %v", err)
			require.NoErrorf(t, actual.Load(fs), "Failed to load config file: %v", err)
			require.Equalf(t, testCase.envValue, actual.Keyspace.ReplicationClass, "Failed to load environment variable into configs")
			require.NoErrorf(t, os.Unsetenv(keyspaceKey), "Failed to unset environment variable set for test")
		})
	}
}

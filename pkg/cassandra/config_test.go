package cassandra

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

func TestCassandraConfigs_Load(t *testing.T) {
	keyspaceKey := fmt.Sprintf("%s_KEYSPACE.REPLICATION_CLASS", constants.GetCassandraPrefix())

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
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetCassandraFileName(), []byte(testCase.input), 0644), "Failed to write in memory file")

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
			t.Setenv(keyspaceKey, testCase.envValue)
			require.NoErrorf(t, actual.Load(fs), "Failed to load constants file: %v", err)
			require.Equalf(t, testCase.envValue, actual.Keyspace.ReplicationClass, "Failed to load environment variable into configs")
		})
	}
}

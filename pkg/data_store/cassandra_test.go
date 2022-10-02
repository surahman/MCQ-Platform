package data_store

import (
	"reflect"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/config"
	"github.com/surahman/mcq-platform/pkg/logger"
)

var configTestData = config.CassandraConfigTestData()

func TestNewCassandra(t *testing.T) {
	log, _ := logger.NewTestLogger()
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(config.GetEtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, config.GetEtcDir()+config.GetCassandraFileName(),
		[]byte(configTestData["valid"]), 0644), "Failed to write in memory file")

	testCases := []struct {
		name      string
		fs        *afero.Fs
		log       *logger.Logger
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"Invalid file system and logger",
			nil,
			nil,
			require.Error,
			require.Nil,
		}, {
			"Invalid file system",
			nil,
			log,
			require.Error,
			require.Nil,
		}, {
			"Invalid logger",
			&fs,
			nil,
			require.Error,
			require.Nil,
		}, {
			"Valid",
			&fs,
			log,
			require.NoError,
			require.NotNil,
		},
		// ----- test cases end ----- //
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cassandra, err := NewCassandra(testCase.fs, testCase.log)
			testCase.expectErr(t, err)
			testCase.expectNil(t, cassandra)
		})
	}
}

func TestNewCassandraImpl(t *testing.T) {
	testCases := []struct {
		name      string
		fullPath  string
		fileName  string
		input     string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"File found",
			config.GetEtcDir(),
			config.GetCassandraFileName(),
			configTestData["valid"],
			require.NoError,
			require.NotNil,
		}, {
			"File not found",
			config.GetEtcDir(),
			"wrong_file_name.yaml",
			configTestData["valid"],
			require.Error,
			require.Nil,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(testCase.fullPath, 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, testCase.fullPath+testCase.fileName, []byte(testCase.input), 0644), "Failed to write in memory file")

			testLogger, err := logger.NewTestLogger()
			require.NoError(t, err, "Failed to create Zap logger for testing.")
			c, err := NewCassandra(&fs, testLogger)
			testCase.expectErr(t, err)
			testCase.expectNil(t, c)
		})
	}
}

func TestCassandraImpl_Execute(t *testing.T) {
	type testType struct {
		key string
		val string
	}

	input := &testType{key: "key", val: "value"}
	fn := func(conn Cassandra, params any) (any, error) {
		casted := params.(*testType)
		return casted, nil
	}

	result, err := connection.db.Execute(fn, input)
	require.NoError(t, err)
	require.Equal(t, reflect.TypeOf(input), reflect.TypeOf(result.(*testType)))
}

func TestCassandra_Open(t *testing.T) {
	cassandra, err := getTestConfiguration()
	require.NoError(t, err, "Failed to open a connection to the cluster")
	cassandra.conf.Keyspace.Name = integrationKeyspace

	require.NoError(t, cassandra.Open(), "Failed to open a connection to the cluster")

	require.Error(t, cassandra.Open(), "Attempt to leak connection pool")
	require.NoError(t, cassandra.Close(), "Failed to close connection")
}

func TestCassandra_Close(t *testing.T) {
	var cassandra *CassandraImpl
	var err error

	cassandra, err = getTestConfiguration()
	require.NoError(t, err, "Failed to get test configuration.")
	cassandra.conf.Keyspace.Name = integrationKeyspace

	require.Error(t, cassandra.Close(), "Should return an error when a connection is not initially established")

	require.NoError(t, cassandra.Open(), "Should establish a connection")
	require.NoError(t, cassandra.Close(), "Should close an established connection")

	require.Error(t, cassandra.Close(), "Should return an error when a connection was closed")
}

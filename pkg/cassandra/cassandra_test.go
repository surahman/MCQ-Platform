package cassandra

import (
	"reflect"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/logger"
)

func TestNewCassandra(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetCassandraFileName(),
		[]byte(cassandraConfigTestData["valid"]), 0644), "Failed to write in memory file")

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
			zapLogger,
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
			zapLogger,
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
		fileName  string
		input     string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"File found",
			constants.GetCassandraFileName(),
			cassandraConfigTestData["valid"],
			require.NoError,
			require.NotNil,
		}, {
			"File not found",
			"wrong_file_name.yaml",
			cassandraConfigTestData["valid"],
			require.Error,
			require.Nil,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+testCase.fileName, []byte(testCase.input), 0644), "Failed to write in memory file")

			c, err := NewCassandra(&fs, zapLogger)
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

	// Configure mock filesystem.
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetCassandraFileName(), []byte(cassandraConfigTestData["valid"]), 0644), "Failed to write in memory file")

	db, err := NewCassandra(&fs, zapLogger)
	require.NoError(t, err, "failed to create test db object")
	require.NotNil(t, db, "failed to create test db connection")

	result, err := db.Execute(fn, input)
	require.NoError(t, err)
	require.Equal(t, reflect.TypeOf(input), reflect.TypeOf(result.(*testType)))
}

func TestCassandra_Open(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	cassandra, err := getTestConfiguration()
	require.NoError(t, err, "Failed to open a connection to the cluster")
	cassandra.conf.Keyspace.Name = integrationKeyspace

	require.NoError(t, cassandra.Open(), "Failed to open a connection to the cluster")

	require.Error(t, cassandra.Open(), "Attempt to leak connection pool")
	require.NoError(t, cassandra.Close(), "Failed to close connection")
}

func TestCassandra_Close(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	var cassandra *cassandraImpl
	var err error

	cassandra, err = getTestConfiguration()
	require.NoError(t, err, "Failed to get test configuration.")
	cassandra.conf.Keyspace.Name = integrationKeyspace

	require.Error(t, cassandra.Close(), "Should return an error when a connection is not initially established")

	require.NoError(t, cassandra.Open(), "Should establish a connection")
	require.NoError(t, cassandra.Close(), "Should close an established connection")

	require.Error(t, cassandra.Close(), "Should return an error when a connection was closed")
}

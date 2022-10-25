package redis

import (
	"testing"

	"github.com/go-redis/redis/v9"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/logger"
)

func TestNewRedisImpl(t *testing.T) {
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
			constants.GetRedisFileName(),
			redisConfigTestData["valid"],
			require.NoError,
			require.NotNil,
		}, {
			"File not found",
			"wrong_file_name.yaml",
			redisConfigTestData["valid"],
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

			c, err := newRedisImpl(&fs, zapLogger)
			testCase.expectErr(t, err)
			testCase.expectNil(t, c)
		})
	}
}

func TestNewRedis(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetRedisFileName(),
		[]byte(redisConfigTestData["valid"]), 0644), "Failed to write in memory file")

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
			cassandra, err := NewRedis(testCase.fs, testCase.log)
			testCase.expectErr(t, err)
			testCase.expectNil(t, cassandra)
		})
	}
}

func TestVerifySession(t *testing.T) {
	nilConnection := redisImpl{redisDb: nil}
	require.Error(t, nilConnection.verifySession(), "nil connection should return an error")

	badConnection := redisImpl{redisDb: redis.NewClusterClient(&redis.ClusterOptions{})}
	require.Error(t, badConnection.verifySession(), "verifying a not open connection should return an error")
}
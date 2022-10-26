package redis

import (
	"testing"

	"github.com/go-redis/redis/v9"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/logger"
	"gopkg.in/yaml.v3"
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

func TestRedisImpl_Open(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Ping failure.
	noNodes := redisImpl{conf: &config{}, logger: zapLogger}
	err := noNodes.Open()
	require.Error(t, err, "connection should fail to ping the cluster")
	require.Contains(t, err.Error(), "no nodes", "error should contain information on no nodes")

	// Connection success.
	conf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData["valid"]), &conf), "failed to prepare test config")
	testRedis := redisImpl{conf: &conf, logger: zapLogger}
	require.NoError(t, testRedis.Open(), "failed to create new cluster connection")

	// Leaked connection check.
	require.Error(t, testRedis.Open(), "leaking a connection should raise an error")
}

func TestRedisImpl_HealthCheck(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Open unhealthy connection, ignore error, and run check.
	unhealthyConf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData["valid"]), &unhealthyConf), "failed to prepare unhealthy config")
	unhealthyConf.Connection.Addrs = []string{"127.0.0.1:8000", "127.0.0.1:8001"}
	unhealthy := redisImpl{conf: &unhealthyConf, logger: zapLogger}
	require.Error(t, unhealthy.Open(), "opening a connection to bad endpoints should fail")
	err := unhealthy.HealthCheck()
	require.Error(t, err, "unhealthy healthcheck failed")
	require.Contains(t, err.Error(), "connection refused", "error is not about a bad connection")

	// Open healthy connection, ignore error, and run check.
	healthyConf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData["valid"]), &healthyConf), "failed to prepare healthy config")
	healthy := redisImpl{conf: &healthyConf, logger: zapLogger}
	require.NoError(t, healthy.Open(), "opening a connection to good endpoints should not fail")
	err = healthy.HealthCheck()
	require.NoError(t, err, "healthy healthcheck failed")
}

package redis

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
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
			redisConfigTestData["test_suite"],
			require.NoError,
			require.NotNil,
		}, {
			"File not found",
			"wrong_file_name.yaml",
			redisConfigTestData["test_suite"],
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
		[]byte(redisConfigTestData["test_suite"]), 0644), "Failed to write in memory file")

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
	noNodesCfg := &config{}
	noNodesCfg.Connection.MaxConnAttempts = 1
	noNodes := redisImpl{conf: noNodesCfg, logger: zapLogger}
	err := noNodes.Open()
	require.Error(t, err, "connection should fail to ping the cluster")
	require.Contains(t, err.Error(), "no nodes", "error should contain information on no nodes")

	// Connection success.
	conf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData["test_suite"]), &conf), "failed to prepare test config")
	testRedis := redisImpl{conf: &conf, logger: zapLogger}
	require.NoError(t, testRedis.Open(), "failed to create new cluster connection")

	// Leaked connection check.
	require.Error(t, testRedis.Open(), "leaking a connection should raise an error")
}

func TestRedisImpl_Close(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Ping failure.
	noNodesCfg := &config{}
	noNodesCfg.Connection.MaxConnAttempts = 1
	noNodes := redisImpl{conf: noNodesCfg, logger: zapLogger}
	err := noNodes.Close()
	require.Error(t, err, "connection should fail to ping the cluster")
	require.Contains(t, err.Error(), "no session", "error should contain information on no nodes")

	// Connection success.
	conf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData["test_suite"]), &conf), "failed to prepare test config")
	testRedis := redisImpl{conf: &conf, logger: zapLogger}
	require.NoError(t, testRedis.Open(), "failed to open cluster connection for test")
	require.NoError(t, testRedis.Close(), "failed to close cluster connection")

	// Leaked connection check.
	require.Error(t, testRedis.Close(), "closing a closed cluster connection should raise an error")
}

func TestRedisImpl_Healthcheck(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Open unhealthy connection, ignore error, and run check.
	unhealthyConf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData["test_suite"]), &unhealthyConf), "failed to prepare unhealthy config")
	unhealthyConf.Connection.Addrs = []string{"127.0.0.1:8000", "127.0.0.1:8001"}
	unhealthy := redisImpl{conf: &unhealthyConf, logger: zapLogger}
	require.Error(t, unhealthy.Open(), "opening a connection to bad endpoints should fail")
	err := unhealthy.Healthcheck()
	require.Error(t, err, "unhealthy healthcheck failed")
	require.Contains(t, err.Error(), "connection refused", "error is not about a bad connection")

	// Open healthy connection, ignore error, and run check.
	healthyConf := config{}
	require.NoError(t, yaml.Unmarshal([]byte(redisConfigTestData["test_suite"]), &healthyConf), "failed to prepare healthy config")
	healthy := redisImpl{conf: &healthyConf, logger: zapLogger}
	require.NoError(t, healthy.Open(), "opening a connection to good endpoints should not fail")
	err = healthy.Healthcheck()
	require.NoError(t, err, "healthy healthcheck failed")
}

func TestRedisImpl_Set_Get_Del(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}
	// Lock connection to Redis cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()

	testCases := []struct {
		name string
		quiz *model_cassandra.Quiz
	}{
		// ----- test cases start ----- //
		{
			name: "myPubQuiz",
			quiz: quizzesTestData["myPubQuiz"],
		}, {
			name: "providedPubQuiz",
			quiz: quizzesTestData["providedPubQuiz"],
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			quizId := testCase.quiz.QuizID.String() + constants.GetIntegrationTestKeyspaceSuffix()

			// Write to Redis.
			require.NoError(t, connection.db.Set(quizId, testCase.quiz), "failed to write to Redis")
			time.Sleep(time.Second) // Allow cache propagation.

			// Get data and validate it.
			retrievedQuiz := model_cassandra.Quiz{}
			err := connection.db.Get(quizId, &retrievedQuiz)
			require.NoError(t, err, "failed to retrieve data from Redis")
			require.True(t, reflect.DeepEqual(*testCase.quiz, retrievedQuiz), "retrieved quiz does not match expected")

			// Remove data from cluster.
			require.NoError(t, connection.db.Del(quizId), "failed to remove quiz from Redis cluster")
			time.Sleep(time.Second) // Allow cache propagation.

			// Check to see if data has been removed.
			deletedQuiz := model_cassandra.Quiz{}
			err = connection.db.Get(quizId, &deletedQuiz)
			require.Nil(t, deletedQuiz.QuizCore, "returned data from a deleted record should be nil")
			require.Error(t, err, "deleted record should not be found on redis cluster")

			// Double-delete data.
			require.Error(t, connection.db.Del(quizId), "removing a nonexistent quiz from Redis cluster should fail")
		})
	}
}

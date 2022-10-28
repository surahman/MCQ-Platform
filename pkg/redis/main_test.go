package redis

import (
	"flag"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// testConnection is the connection pool to the Redis cluster. The mutex is used for sequential test execution.
type testConnection struct {
	db Redis        // Test database connection.
	mu sync.RWMutex // Mutex to enforce sequential test execution.
}

// redisConfigTestData is a map of Redis configuration test data.
var redisConfigTestData = configTestData()

// quizzesTestData is a map of Quiz data to be used with the Redis cluster.
var quizzesTestData = cassandra.GetTestQuizzes()

// connection pool to Redis cluster.
var connection testConnection

// zapLogger is the Zap logger used strictly for the test suite in this package.
var zapLogger *logger.Logger

func TestMain(m *testing.M) {
	// Parse commandline flags to check for short tests.
	flag.Parse()

	var err error
	// Configure logger.
	if zapLogger, err = logger.NewTestLogger(); err != nil {
		log.Printf("Test suite logger setup failed: %v\n", err)
		os.Exit(1)
	}

	// Setup test space.
	if err = setup(); err != nil {
		zapLogger.Error("Test suite setup failure", zap.Error(err))
		os.Exit(1)
	}

	// Run test suite.
	exitCode := m.Run()

	// Cleanup test space.
	if err = tearDown(); err != nil {
		zapLogger.Error("Test suite teardown failure:", zap.Error(err))
		os.Exit(1)
	}
	os.Exit(exitCode)
}

// setup will configure the connection to the test clusters keyspace.
func setup() (err error) {
	if testing.Short() {
		zapLogger.Warn("Short test: Skipping Redis integration tests")
		return
	}

	conf := config{}
	if err = yaml.Unmarshal([]byte(redisConfigTestData["valid"]), &conf); err != nil {
		return
	}
	connection.db = &redisImpl{conf: &conf, logger: zapLogger}
	if err = connection.db.Open(); err != nil {
		return
	}

	return
}

// tearDown will delete the test clusters keyspace.
func tearDown() (err error) {
	if !testing.Short() {
		return // Close test db connection.
	}
	return
}

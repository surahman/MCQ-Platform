package redis

import (
	"flag"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
)

// testConnection is the connection pool to the Redis cluster. The mutex is used for sequential test execution.
type testConnection struct {
	db Redis        // Test database connection.
	mu sync.RWMutex // Mutex to enforce sequential test execution.
}

// redisConfigTestData is a map of Redis configuration test data.
var redisConfigTestData = configTestData()

// connection pool to Cassandra cluster.
var connection testConnection

// zapLogger is the Zap logger used strictly for the test suite in this package.
var zapLogger *logger.Logger

// integrationKeyspace is the name of the keyspace in which testing is conducted.
var integrationKeyspace string

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
		zapLogger.Warn("Short test: Skipping Cassandra integration tests")
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

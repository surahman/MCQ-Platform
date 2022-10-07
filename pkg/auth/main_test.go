package auth

import (
	"log"
	"os"
	"testing"

	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
)

// authConfigTestData is a map of Authentication configuration test data.
var authConfigTestData = configTestData()

// zapLogger is the Zap logger used strictly for the test suite in this package.
var zapLogger *logger.Logger

func TestMain(m *testing.M) {
	var err error
	// Configure logger.
	if zapLogger, err = logger.NewTestLogger(); err != nil {
		log.Printf("Test suite logger setup failed: %v\n", err)
		os.Exit(1)
	}

	// Setup test space.
	if err := setup(); err != nil {
		zapLogger.Error("Test suite setup failure", zap.Error(err))
		os.Exit(1)
	}

	// Run test suite.
	exitCode := m.Run()

	// Cleanup test space.
	if err := tearDown(); err != nil {
		zapLogger.Error("Test suite teardown failure:", zap.Error(err))
		os.Exit(1)
	}
	os.Exit(exitCode)
}

// setup will configure the connection to the test clusters keyspace.
func setup() (err error) {
	return
}

// tearDown will delete the test clusters keyspace.
func tearDown() (err error) {
	return
}

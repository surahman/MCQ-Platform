package http_handlers

import (
	"log"
	"os"
	"testing"

	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
)

// zapLogger is the Zap logger used strictly for the test suite in this package.
var zapLogger *logger.Logger

// testUserData is the test user account data.
var testUserData = cassandra.GetTestUsers()

// testQuizData is the test quiz data.
var testQuizData = cassandra.GetTestQuizzes()

func TestMain(m *testing.M) {
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

// setup will configure the auth test object.
func setup() (err error) {
	return
}

// tearDown will delete the test clusters keyspace.
func tearDown() (err error) {
	return
}

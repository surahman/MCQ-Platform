package http_handlers

import (
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
)

// zapLogger is the Zap logger used strictly for the test suite in this package.
var zapLogger *logger.Logger

// testUserData is the test user account data.
var testUserData = cassandra.GetTestUsers()

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

// getRouter creates a gin router testing instance.
func getRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}

// mockAuthData is the parameter data for Auth mocking that is used in the test grid.
type mockAuthData struct {
	inputParam1 string
	inputParam2 string
	outputParam any
	outputErr   error
	times       int
}

// mockCassandraData is the parameter data for Cassandra mocking that is used in the test grid.
type mockCassandraData struct {
	inputFunc   func(cassandra.Cassandra, any) (any, error)
	inputParam  any
	outputParam any
	outputErr   error
	times       int
}
package auth

import (
	"log"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
)

// testAuth is the Authorization object.
var testAuth Auth

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
	if testAuth, err = getTestConfiguration(); err != nil {
		return
	}
	return
}

// tearDown will delete the test clusters keyspace.
func tearDown() (err error) {
	return
}

// getTestConfiguration creates a cluster configuration for testing.
func getTestConfiguration() (auth *authImpl, err error) {
	// Setup mock filesystem.
	fs := afero.NewMemMapFs()
	if err = fs.MkdirAll(constants.GetEtcDir(), 0644); err != nil {
		return
	}
	if err = afero.WriteFile(fs, constants.GetEtcDir()+constants.GetAuthFileName(), []byte(authConfigTestData["valid"]), 0644); err != nil {
		return
	}

	// Load Cassandra configurations.
	if auth, err = newAuthImpl(&fs, zapLogger); err != nil {
		return
	}

	return
}

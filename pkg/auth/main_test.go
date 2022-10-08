package auth

import (
	"log"
	"os"
	"testing"

	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
)

// testAuth is the Authorization object.
var testAuth Auth

// authConfigTestData is a map of Authentication configuration test data.
var authConfigTestData = configTestData()

// expirationDuration is the time in seconds that a JWT will be valid for.
var expirationDuration int64 = 10

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

// setup will configure the auth test object.
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

// getTestConfiguration creates an Auth configuration for testing.
func getTestConfiguration() (auth *authImpl, err error) {
	auth = &authImpl{
		conf:   &config{},
		logger: zapLogger,
	}
	auth.conf.JWTConfig.Key = "encryption key for test suite"
	auth.conf.JWTConfig.Issuer = "issuer for test suite"
	auth.conf.JWTConfig.ExpirationDuration = expirationDuration
	auth.conf.General.BcryptCost = 4

	return
}

package auth

import (
	"errors"

	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
)

// Mock Auth interface stub generation.
//go:generate mockgen -destination=../mocks/mock_auth.go -package=mocks github.com/surahman/mcq-platform/pkg/auth Auth

// Auth is the interface through which the cluster can be accessed. Created to support mock testing.
type Auth interface {
}

// Check to ensure the Cassandra interface has been implemented.
var _ Auth = &authImpl{}

// authImpl implements the Auth interface and contains the logic for authorization functionality.
type authImpl struct {
	conf   *config
	logger *logger.Logger
}

// NewAuth will create a new Authorization configuration by loading it.
func NewAuth(fs *afero.Fs, logger *logger.Logger) (Auth, error) {
	if fs == nil || logger == nil {
		return nil, errors.New("nil file system of logger supplied")
	}
	return newAuthImpl(fs, logger)
}

// newAuthImpl will create a new CassandraImpl configuration and load it from disk.
func newAuthImpl(fs *afero.Fs, logger *logger.Logger) (c *authImpl, err error) {
	c = &authImpl{conf: newConfig(), logger: logger}
	if err = c.conf.Load(*fs); err != nil {
		c.logger.Error("failed to load Authorization configurations from disk", zap.Error(err))
		return nil, err
	}
	return
}

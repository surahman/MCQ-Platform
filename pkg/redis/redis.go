package redis

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis/v9"
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
)

// Mock Redis interface stub generation.
//go:generate mockgen -destination=../mocks/mock_redis.go -package=mocks github.com/surahman/mcq-platform/pkg/redis Redis

// Redis is the interface through which the cluster can be accessed. Created to support mock testing.
type Redis interface {
	// Open will create a connection pool and establish a connection to the cache backend.
	Open() error

	// Close will shut down the connection pool and ensure that the connection to the database backend is terminated correctly.
	Close() error
}

// Check to ensure the Redis interface has been implemented.
var _ Redis = &redisImpl{}

// redisImpl implements the Redis interface and contains the logic to interface with the cluster.
type redisImpl struct {
	conf    *config
	redisDb *redis.ClusterClient
	logger  *logger.Logger
}

// NewRedis will create a new Redis configuration by loading it.
func NewRedis(fs *afero.Fs, logger *logger.Logger) (Redis, error) {
	if fs == nil || logger == nil {
		return nil, errors.New("nil file system of logger supplied")
	}
	return newRedisImpl(fs, logger)
}

// newRedisImpl will create a new redisImpl configuration and load it from disk.
func newRedisImpl(fs *afero.Fs, logger *logger.Logger) (c *redisImpl, err error) {
	c = &redisImpl{conf: newConfig(), logger: logger}
	if err = c.conf.Load(*fs); err != nil {
		c.logger.Error("failed to load Redis configurations from disk", zap.Error(err))
		return nil, err
	}
	return
}

// Open will establish a connection to the Redis cache backend.
func (r *redisImpl) Open() (err error) {
	return fmt.Errorf("not implemented")
}

// Close will terminate a connection to the Redis cache backend.
func (r *redisImpl) Close() (err error) {
	return r.redisDb.Close()
}

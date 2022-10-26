package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v9"
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
)

// Mock Redis interface stub generation.
//go:generate mockgen -destination=../mocks/mock_redis.go -package=mocks github.com/surahman/mcq-platform/pkg/redis Redis

// Redis is the interface through which the cluster can be accessed. Created to support mock testing.
type Redis interface {
	// Open will create a connection pool and establish a connection to the cache cluster.
	Open() error

	// Close will shut down the connection pool and ensure that the connection to the cache cluster is terminated correctly.
	Close() error

	// HealthCheck will ping all the nodes in the cluster to see if all the shards are reachable.
	HealthCheck() error
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
		return nil, errors.New("nil file system or logger supplied")
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

// verifySession will check to see if a session is established.
func (r *redisImpl) verifySession() error {
	if r.redisDb == nil || r.redisDb.Ping(context.Background()).Err() != nil {
		return errors.New("no session established")
	}
	return nil
}

// Open will establish a connection to the Redis cache cluster.
func (r *redisImpl) Open() (err error) {
	// Stop connection leaks.
	if err = r.verifySession(); err == nil {
		r.logger.Warn("session to cluster is already established")
		return errors.New("session to cluster is already established")
	}

	r.redisDb = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:          r.conf.Connection.Addrs,
		MaxRedirects:   r.conf.Connection.MaxRedirects,
		ReadOnly:       r.conf.Connection.ReadOnly,
		RouteByLatency: r.conf.Connection.RouteByLatency,
		Password:       r.conf.Authentication.Password,
		MaxRetries:     r.conf.Connection.MaxRetries,
		PoolSize:       r.conf.Connection.PoolSize,
		MinIdleConns:   r.conf.Connection.MinIdleConns,
	})

	if err = r.redisDb.Ping(context.Background()).Err(); err != nil {
		r.logger.Error("failed to establish redis cluster connection", zap.Error(err))
		return
	}

	return nil
}

// Close will terminate a connection to the Redis cache cluster.
func (r *redisImpl) Close() (err error) {
	return r.redisDb.Close()
}

// HealthCheck will iterate through all the data shards and attempt to ping them to ensure they are all reachable.
func (r *redisImpl) HealthCheck() (err error) {
	err = r.redisDb.ForEachShard(context.Background(), func(ctx context.Context, shard *redis.Client) error {
		return shard.Ping(ctx).Err()
	})
	return
}

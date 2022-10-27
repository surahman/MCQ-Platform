package redis

import (
	"context"
	"errors"
	"time"

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

	// Healthcheck will ping all the nodes in the cluster to see if all the shards are reachable.
	Healthcheck() error

	// Set will place a key with a given value in the cluster with a TTL, if specified in the configurations.
	Set(string, any) error

	// Get will retrieve a value associated with a provided key.
	Get(string) ([]byte, error)

	// Del will remove all keys provided as a set of keys.
	Del(...string) error
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
	// Check for an open connection.
	if err = r.verifySession(); err != nil {
		msg := "no session to cluster established to close"
		r.logger.Warn(msg)
		return errors.New(msg)
	}
	return r.redisDb.Close()
}

// Healthcheck will iterate through all the data shards and attempt to ping them to ensure they are all reachable.
func (r *redisImpl) Healthcheck() (err error) {
	err = r.redisDb.ForEachShard(context.Background(), func(ctx context.Context, shard *redis.Client) error {
		return shard.Ping(ctx).Err()
	})
	return
}

// Set will place a key with a given value in the cluster with a TTL, if specified in the configurations.
func (r *redisImpl) Set(key string, value any) (err error) {
	if err = r.redisDb.Set(context.Background(), key, value, time.Duration(r.conf.Data.TTL)).Err(); err != nil {
		r.logger.Error("failed to place item in Redis cache", zap.String("key", key), zap.Error(err))
		return
	}
	return
}

// Get will retrieve a value associated with a provided key.
func (r *redisImpl) Get(key string) (val []byte, err error) {
	var response string
	if response, err = r.redisDb.Get(context.Background(), key).Result(); err != nil {
		return
	}
	return []byte(response), err
}

// Del will remove all keys provided as a list of keys.
func (r *redisImpl) Del(keys ...string) (err error) {
	for _, key := range keys {
		var status int64
		status, err = r.redisDb.Del(context.Background(), key).Result()
		if err != nil {
			r.logger.Error("failed to evict item from Redis cache", zap.String("key", key), zap.Error(err))
			return
		}
		if status == 0 {
			err = errors.New("unable to locate key on Redis cluster")
			r.logger.Warn("failed to evict item from Redis cache", zap.String("key", key), zap.Error(err))
			return
		}
	}
	return
}

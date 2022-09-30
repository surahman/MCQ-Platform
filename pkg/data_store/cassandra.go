package data_store

import (
	"errors"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/config"
	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
)

// Cassandra is the interface through which the cluster can be accessed. Created to support mock testing.
type Cassandra interface {
	Open() error
	Close() error
	Execute(func(*CassandraImpl, any) (any, error), any) (any, error)
}

// Check to ensure the Cassandra interface has been implemented.
var _ Cassandra = &CassandraImpl{}

// CassandraImpl implements the Cassandra interface and contains the logic to interface with the cluster.
type CassandraImpl struct {
	conf    *config.CassandraConfig
	session *gocql.Session
	logger  *logger.Logger
}

// NewCassandra will create a new Cassandra configuration by loading it.
func NewCassandra(fs *afero.Fs, logger *logger.Logger) (Cassandra, error) {
	if fs == nil || logger == nil {
		return nil, errors.New("nil file system of logger supplied")
	}
	return newCassandraImpl(fs, logger)
}

// newCassandraImpl will create a new CassandraImpl configuration and load it from disk.
func newCassandraImpl(fs *afero.Fs, logger *logger.Logger) (c *CassandraImpl, err error) {
	c = &CassandraImpl{conf: config.NewCassandraConfig(), logger: logger}
	if err = c.conf.Load(*fs); err != nil {
		c.logger.Error("failed to load Cassandra config from disk", zap.Error(err))
		return nil, err
	}
	return
}

// Open will start a database connection pool and establish a connection.
func (c *CassandraImpl) Open() (err error) {
	// Stop connection leaks.
	if err = c.verifySession(); err == nil {
		c.logger.Warn("session to cluster is already established")
		return errors.New("session to cluster is already established")
	}

	// Configure connection.
	var cluster *gocql.ClusterConfig
	cluster, err = c.configureCluster()
	if err != nil {
		c.logger.Error("failed to configure connection to Cassandra cluster", zap.Error(err))
		return
	}

	// Session connection pool.
	if err = c.createSessionRetry(cluster); err != nil {
		return
	}

	return
}

// Close the CassandraImpl cluster connection.
func (c *CassandraImpl) Close() error {
	if err := c.verifySession(); err != nil {
		return err
	}
	c.session.Close()
	return nil
}

// Execute wraps the methods that create, read, update, and delete records from tables on the Cassandra cluster.
func (c *CassandraImpl) Execute(request func(*CassandraImpl, any) (any, error), params any) (any, error) {
	return request(c, params)
}

// configureCluster will configure the settings for the Cassandra cluster.
func (c *CassandraImpl) configureCluster() (cluster *gocql.ClusterConfig, err error) {
	cluster = gocql.NewCluster(c.conf.Connection.ClusterIP...)
	cluster.ProtoVersion = c.conf.Connection.ProtoVersion
	cluster.ConnectTimeout = time.Duration(c.conf.Connection.Timeout) * time.Second
	cluster.Timeout = time.Duration(c.conf.Connection.Timeout) * time.Second
	if cluster.Consistency, err = gocql.ParseConsistencyWrapper(c.conf.Connection.Consistency); err != nil {
		c.logger.Error("failed to parse Cassandra consistency level provided in user configs", zap.Error(err))
		return nil, err
	}
	cluster.Keyspace = c.conf.Keyspace.Name

	// Configure authentication.
	cluster.Authenticator = gocql.PasswordAuthenticator{Username: c.conf.Authentication.Username, Password: c.conf.Authentication.Password}

	return
}

// verifySession will check to see if a session is established.
func (c *CassandraImpl) verifySession() error {
	if c.session == nil || c.session.Closed() {
		return errors.New("no session established")
	}
	return nil
}

// createSessionRetry will attempt to open the connection a few times stop on the first success or fail after the last one.
func (c *CassandraImpl) createSessionRetry(cluster *gocql.ClusterConfig) (err error) {
	maxAttempts := config.GetCassandraMaxConnectRetries()
	for attempt := 0; attempt <= maxAttempts; attempt++ {
		c.logger.Info("Attempting to connect to Cassandra cluster...", zap.String("attempt", strconv.Itoa(attempt)))
		if c.session, err = cluster.CreateSession(); err == nil {
			break
		}
	}
	if err != nil {
		c.logger.Error("unable to establish connection to Cassandra cluster", zap.Error(err))
	}
	return
}

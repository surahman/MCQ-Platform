package cassandra

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/blake2b"
)

// Mock Cassandra interface stub generation.
//go:generate mockgen -destination=../mocks/mock_cassandra.go -package=mocks github.com/surahman/mcq-platform/pkg/cassandra Cassandra

// Cassandra is the interface through which the cluster can be accessed. Created to support mock testing.
type Cassandra interface {
	// Open will create a connection pool and establish a connection to the database backend.
	Open() error

	// Close will shut down the connection pool and ensure that the connection to the database backend is terminated correctly.
	Close() error

	// Execute will execute statements or run a lightweight transaction on the database backend, leveraging the connection pool.
	Execute(func(Cassandra, any) (any, error), any) (any, error)
}

// Check to ensure the Cassandra interface has been implemented.
var _ Cassandra = &cassandraImpl{}

// CassandraImpl implements the Cassandra interface and contains the logic to interface with the cluster.
type cassandraImpl struct {
	conf    *config
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
func newCassandraImpl(fs *afero.Fs, logger *logger.Logger) (c *cassandraImpl, err error) {
	c = &cassandraImpl{conf: newConfig(), logger: logger}
	if err = c.conf.Load(*fs); err != nil {
		c.logger.Error("failed to load Cassandra configurations from disk", zap.Error(err))
		return nil, err
	}
	return
}

// Open will start a database connection pool and establish a connection.
func (c *cassandraImpl) Open() (err error) {
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
func (c *cassandraImpl) Close() error {
	if err := c.verifySession(); err != nil {
		return err
	}
	c.session.Close()
	return nil
}

// Execute wraps the methods that create, read, update, and delete records from tables on the Cassandra cluster.
func (c *cassandraImpl) Execute(request func(Cassandra, any) (any, error), params any) (any, error) {
	return request(c, params)
}

// configureCluster will configure the settings for the Cassandra cluster.
func (c *cassandraImpl) configureCluster() (cluster *gocql.ClusterConfig, err error) {
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
func (c *cassandraImpl) verifySession() error {
	if c.session == nil || c.session.Closed() {
		return errors.New("no session established")
	}
	return nil
}

// createSessionRetry will attempt to open the connection using binary exponential back-off and stop on the first success or fail after the last one.
func (c *cassandraImpl) createSessionRetry(cluster *gocql.ClusterConfig) (err error) {
	maxAttempts := constants.GetCassandraMaxConnectRetries()
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		waitTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		c.logger.Info(fmt.Sprintf("Attempting connection to Cassandra cluster in %s...", waitTime), zap.String("attempt", strconv.Itoa(attempt)))
		time.Sleep(waitTime)
		if c.session, err = cluster.CreateSession(); err == nil {
			return
		}
	}
	c.logger.Error("unable to establish connection to Cassandra cluster", zap.Error(err))
	return
}

// blake2b256 will create a hash from an input string. This hash is used to create the Account ID for a user.
func blake2b256(data string) string {
	hash := blake2b.Sum256([]byte(data))
	return base64.URLEncoding.EncodeToString(hash[:])
}

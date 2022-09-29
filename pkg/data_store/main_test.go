package data_store

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/gocql/gocql"
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/config"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Setup test space.
	if err := setup(); err != nil {
		log.Printf("Test suite setup failure: %v\n", err)
		os.Exit(1)
	}

	// Run test suite.
	exitCode := m.Run()

	// Cleanup test space.
	if err := tearDown(); err != nil {
		log.Printf("Test suite teardown failure: %v\n", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

// testConnection is the connection pool to the Cassandra cluster. The mutex is used for sequential test execution.
type testConnection struct {
	db         Cassandra // Test database connection.
	sync.Mutex           // Mutex to enforce sequential test execution.
}

// connection pool to Cassandra cluster.
var connection testConnection

// setup will configure the connection to the test clusters keyspace.
func setup() (err error) {
	// Setup mock filesystem.
	fs := afero.NewMemMapFs()
	if err = fs.MkdirAll(config.GetEtcDir(), 0644); err != nil {
		return
	}
	cassandraConf := config.CassandraConfigTestData()["valid"]
	if err = afero.WriteFile(fs, config.GetEtcDir()+config.GetCassandraFileName(), []byte(cassandraConf), 0644); err != nil {
		return
	}

	// Configure logger.
	var zapLogger *logger.Logger
	if zapLogger, err = logger.NewTestLogger(); err != nil {
		return
	}

	// Load Cassandra configurations.
	if connection.db, err = NewCassandra(&fs, zapLogger); err != nil {
		return
	}

	// Create Keyspace for integration test.
	if err = createTestingKeyspace(connection.db.(*CassandraImpl)); err != nil {
		return
	}

	// Open connection to cluster in integration test keyspace.

	// Migrate schema to integration test keyspace.

	return
}

// tearDown will delete the test clusters keyspace.
func tearDown() error {
	return connection.db.Close()
}

// createTestingKeyspace will configure and create a fresh Keyspace for integration testing and connect to it.
func createTestingKeyspace(c *CassandraImpl) (err error) {
	if err = c.verifySession(); err == nil {
		return errors.New("session to Cassandra already established")
	}

	// Configure cluster configs for integration test.
	var cluster *gocql.ClusterConfig
	if cluster, err = c.configureCluster(); err != nil {
		return err
	}

	// Connection scoped to cluster wide to create integration test keyspace.
	cluster.Keyspace = ""
	integrationKeyspace := c.conf.Keyspace.Name + config.GetIntegrationTestKeyspaceSuffix()

	// Create keyspace connection pool.
	if err = c.createSessionRetry(cluster); err != nil {
		c.logger.Error("unable to establish connection to Cassandra cluster", zap.Error(err))
		return
	}

	// Drop and create a fresh keyspace.
	if err = c.session.Query(fmt.Sprintf("DROP KEYSPACE IF EXISTS %s;", integrationKeyspace)).Exec(); err != nil {
		c.logger.Error("failed to drop integration test keyspace", zap.Error(err))
		return
	}
	if err = c.session.Query(
		fmt.Sprintf(model_cassandra.CreateKeyspace, integrationKeyspace, c.conf.Keyspace.ReplicationClass, 1)).Exec(); err != nil {
		c.logger.Error("failed to create integration test keyspace", zap.Error(err))
	}
	c.logger.Info("fresh integration test keyspace created", zap.String("name", integrationKeyspace))

	// Close connection to create keyspace and open keyspace scoped connection.
	c.session.Close()
	cluster.Keyspace = integrationKeyspace
	if err = c.createSessionRetry(cluster); err != nil {
		c.logger.Error("unable to establish connection to Cassandra cluster scoped to integration test keyspace", zap.Error(err))
		return
	}
	c.logger.Info("connected to cluster and scoped to integration test keyspace", zap.String("name", integrationKeyspace))

	// Create users, quizzes, and responses tables.
	createTablesWg := sync.WaitGroup{}
	createTablesWg.Add(3)
	errorsChan := make(chan error, 3)

	go createUsersTable(c, errorsChan, &createTablesWg)
	go createQuizzesTable(c, errorsChan, &createTablesWg)
	go createResponsesTable(c, errorsChan, &createTablesWg)

	createTablesWg.Wait()
	close(errorsChan)
	for err = range errorsChan {
		if err != nil {
			return
		}
	}

	return
}

// createUsersTable will create the users table in the integration test keyspace.
func createUsersTable(c *CassandraImpl, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := c.session.Query(model_cassandra.CreateUsersTable).Exec(); err != nil {
		c.logger.Error("failed to create users table in integration test keyspace", zap.Error(err))
		errors <- err
		return
	}
	c.logger.Info("created users table in integration test keyspace")
}

// createQuizzesTable will create the quizzes table in the integration test keyspace.
func createQuizzesTable(c *CassandraImpl, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := c.session.Query(model_cassandra.CreateQuestionUDT).Exec(); err != nil {
		c.logger.Error("failed to create questions UDT in integration test keyspace", zap.Error(err))
		errors <- err
		return
	}
	c.logger.Info("created questions UDT in integration test keyspace")
	if err := c.session.Query(model_cassandra.CreateQuizzesTable).Exec(); err != nil {
		c.logger.Error("failed to create quizzes table in integration test keyspace", zap.Error(err))
		errors <- err
		return
	}
	c.logger.Info("created quizzed table in integration test keyspace")
}

// createResponsesTable will create the responses table in the integration test keyspace.
func createResponsesTable(c *CassandraImpl, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := c.session.Query(model_cassandra.CreateResponsesTable).Exec(); err != nil {
		c.logger.Error("failed to create responses table in integration test keyspace", zap.Error(err))
		errors <- err
		return
	}
	c.logger.Info("created responses table in integration test keyspace")
	if err := c.session.Query(model_cassandra.CreateResponsesIndex).Exec(); err != nil {
		c.logger.Error("failed to create responses index in integration test keyspace", zap.Error(err))
		errors <- err
		return
	}
	c.logger.Info("created responses index in integration test keyspace")
}

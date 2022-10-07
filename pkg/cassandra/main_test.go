package cassandra

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/gocql/gocql"
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"go.uber.org/zap"
)

// testConnection is the connection pool to the Cassandra cluster. The mutex is used for sequential test execution.
type testConnection struct {
	db Cassandra    // Test database connection.
	mu sync.RWMutex // Mutex to enforce sequential test execution.
}

// cassandraConfigTestData is a map of Cassandra configuration test data.
var cassandraConfigTestData = configTestData()

// connection pool to Cassandra cluster.
var connection testConnection

// zapLogger is the Zap logger used strictly for the test suite in this package.
var zapLogger *logger.Logger

// integrationKeyspace is the name of the keyspace in which testing is conducted.
var integrationKeyspace string

func TestMain(m *testing.M) {
	// Parse commandline flags to check for short tests.
	flag.Parse()

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
	if testing.Short() {
		zapLogger.Warn("Short test: Skipping Cassandra integration tests")
		return
	}
	// Load Cassandra configurations.
	if connection.db, err = getTestConfiguration(); err != nil {
		return
	}

	// Integration test keyspace name.
	integrationKeyspace = connection.db.(*cassandraImpl).conf.Keyspace.Name + constants.GetIntegrationTestKeyspaceSuffix()

	// Create Keyspace for integration test.
	if err = createTestingKeyspace(connection.db.(*cassandraImpl)); err != nil {
		return
	}

	return
}

// tearDown will delete the test clusters keyspace.
func tearDown() (err error) {
	if !testing.Short() {
		return connection.db.Close()
	}
	return
}

// getTestConfiguration creates a cluster configuration for testing.
func getTestConfiguration() (cassandra *cassandraImpl, err error) {
	// If running on a GitHub Actions runner use the default credentials for Cassandra.
	configFileKey := "valid"
	if _, ok := os.LookupEnv(constants.GetGithubCIKey()); ok == true {
		configFileKey = "valid-ci"
		zapLogger.Info("Integration Test running on Github CI runner.")
	}

	// Setup mock filesystem.
	fs := afero.NewMemMapFs()
	if err = fs.MkdirAll(constants.GetEtcDir(), 0644); err != nil {
		return
	}
	if err = afero.WriteFile(fs, constants.GetEtcDir()+constants.GetCassandraFileName(), []byte(cassandraConfigTestData[configFileKey]), 0644); err != nil {
		return
	}

	// Load Cassandra configurations.
	if cassandra, err = newCassandraImpl(&fs, zapLogger); err != nil {
		return
	}

	return
}

// createTestingKeyspace will configure and create a fresh Keyspace for integration testing and connect to it.
func createTestingKeyspace(c *cassandraImpl) (err error) {
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

	// Close connection used to create keyspace and open keyspace scoped connection.
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
func createUsersTable(c *cassandraImpl, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := c.session.Query(model_cassandra.CreateUsersTable).Exec(); err != nil {
		c.logger.Error("failed to create users table in integration test keyspace", zap.Error(err))
		errors <- err
		return
	}
	c.logger.Info("created users table in integration test keyspace")
}

// createQuizzesTable will create the quizzes table in the integration test keyspace.
func createQuizzesTable(c *cassandraImpl, errors chan<- error, wg *sync.WaitGroup) {
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
func createResponsesTable(c *cassandraImpl, errors chan<- error, wg *sync.WaitGroup) {
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

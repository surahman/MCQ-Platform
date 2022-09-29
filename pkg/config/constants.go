package config

const (
	// Configuration file directories
	configEtcDir  = "/etc/MCQ_Platform.conf/"
	configHomeDir = "$HOME/.MCQ_Platform/"

	// Configuration file names
	cassandraConfigFileName = "CassandraConfig.yaml"
	loggerConfigFileName    = "LoggerConfig.yaml"

	// Environment variables
	cassandraPrefix = "CASSANDRA"
	loggerPrefix    = "LOGGER"

	// Misc.
	integrationTestKeyspaceSuffix = "_integration_testing"
	cassandraMaxConnectRetries    = 5
)

// GetEtcDir returns the configuration directory in Etc.
func GetEtcDir() string {
	return configEtcDir
}

// GetHomeDir returns the configuration directory in users home.
func GetHomeDir() string {
	return configHomeDir
}

// GetCassandraFileName returns the Cassandra configuration file name.
func GetCassandraFileName() string {
	return cassandraConfigFileName
}

// GetLoggerFileName returns the Zap logger configuration file name.
func GetLoggerFileName() string {
	return loggerConfigFileName
}

// GetIntegrationTestKeyspaceSuffix is the suffix attached to the clusters keyspace and is used for integration tests.
func GetIntegrationTestKeyspaceSuffix() string {
	return integrationTestKeyspaceSuffix
}

// GetCassandraMaxConnectRetries is the maximum number of attempts to retry connecting to the Cassandra cluster.
func GetCassandraMaxConnectRetries() int {
	return cassandraMaxConnectRetries
}

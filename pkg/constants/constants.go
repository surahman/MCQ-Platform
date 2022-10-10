package constants

const (
	// Configuration file directories
	configEtcDir  = "/etc/MCQ_Platform.conf/"
	configHomeDir = "$HOME/.MCQ_Platform/"

	// Configuration file names
	cassandraConfigFileName = "CassandraConfig.yaml"
	loggerConfigFileName    = "LoggerConfig.yaml"
	authConfigFileName      = "AuthConfig.yaml"
	restConfigFileName      = "HTTPRESTConfig.yaml"

	// Environment variables
	cassandraPrefix = "CASSANDRA"
	loggerPrefix    = "LOGGER"
	githubCIKey     = "GITHUB_ACTIONS_CI"
	authPrefix      = "AUTH"
	restPrefix      = "REST"

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

// GetAuthFileName returns the authentication configuration file name.
func GetAuthFileName() string {
	return authConfigFileName
}

// GetHTTPRESTFileName returns the HTTP REST endpoint configuration file name.
func GetHTTPRESTFileName() string {
	return restConfigFileName
}

// GetCassandraPrefix returns the environment variable prefix for Cassandra.
func GetCassandraPrefix() string {
	return cassandraPrefix
}

// GetLoggerPrefix returns the environment variable prefix for the Zap logger.
func GetLoggerPrefix() string {
	return loggerPrefix
}

// GetAuthPrefix returns the environment variable prefix for authentication.
func GetAuthPrefix() string {
	return authPrefix
}

// GetHTTPRESTPrefix returns the environment variable prefix for the HTTP REST endpoint.
func GetHTTPRESTPrefix() string {
	return restPrefix
}

// GetIntegrationTestKeyspaceSuffix is the suffix attached to the clusters keyspace and is used for integration tests.
func GetIntegrationTestKeyspaceSuffix() string {
	return integrationTestKeyspaceSuffix
}

// GetCassandraMaxConnectRetries is the maximum number of attempts to retry connecting to the Cassandra cluster.
func GetCassandraMaxConnectRetries() int {
	return cassandraMaxConnectRetries
}

// GetGithubCIKey is the key for the environment variable expected to be present in the GH CI runner.
func GetGithubCIKey() string {
	return githubCIKey
}

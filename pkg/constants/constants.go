package constants

const (
	// Configuration file directories
	configEtcDir  = "/etc/MCQ_Platform.conf/"
	configHomeDir = "$HOME/.MCQ_Platform/"
	configBaseDir = "./configs/"

	// Configuration file names
	cassandraConfigFileName = "CassandraConfig.yaml"
	loggerConfigFileName    = "LoggerConfig.yaml"
	authConfigFileName      = "AuthConfig.yaml"
	restConfigFileName      = "HTTPRESTConfig.yaml"
	graphqlConfigFileName   = "GraphQLConfig.yaml"
	redisConfigFileName     = "RedisConfig.yaml"

	// Environment variables
	cassandraPrefix = "CASSANDRA"
	loggerPrefix    = "LOGGER"
	githubCIKey     = "GITHUB_ACTIONS_CI"
	authPrefix      = "AUTH"
	restPrefix      = "REST"
	graphqlPrefix   = "GRAPHQL"
	redisPrefix     = "REDIS"

	// Misc.
	integrationTestKeyspaceSuffix = "_integration_testing"
	deleteUserAccountConfirmation = "I understand the consequences, delete my user account %s"
)

// GetEtcDir returns the configuration directory in Etc.
func GetEtcDir() string {
	return configEtcDir
}

// GetHomeDir returns the configuration directory in users home.
func GetHomeDir() string {
	return configHomeDir
}

// GetBaseDir returns the configuration base directory in the root of the application.
func GetBaseDir() string {
	return configBaseDir
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

// GetGraphQLFileName returns the HTTP GraphQL endpoint configuration file name.
func GetGraphQLFileName() string {
	return graphqlConfigFileName
}

// GetRedisFileName returns the Redis cluster configuration file name.
func GetRedisFileName() string {
	return redisConfigFileName
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

// GetGraphQLPrefix returns the environment variable prefix for the HTTP GraphQL endpoint.
func GetGraphQLPrefix() string {
	return graphqlPrefix
}

// GetRedisPrefix returns the environment variable prefix for the Redis cluster.
func GetRedisPrefix() string {
	return redisPrefix
}

// GetIntegrationTestKeyspaceSuffix is the suffix attached to the clusters keyspace and is used for integration tests.
func GetIntegrationTestKeyspaceSuffix() string {
	return integrationTestKeyspaceSuffix
}

// GetGithubCIKey is the key for the environment variable expected to be present in the GH CI runner.
func GetGithubCIKey() string {
	return githubCIKey
}

// GetDeleteUserAccountConfirmation is the format string template confirmation message used to delete a user account.
func GetDeleteUserAccountConfirmation() string {
	return deleteUserAccountConfirmation
}

package constants

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEtcDir(t *testing.T) {
	require.Equal(t, configEtcDir, GetEtcDir(), "Incorrect etc directory")
}

func TestGetHomeDir(t *testing.T) {
	require.Equal(t, configHomeDir, GetHomeDir(), "Incorrect home directory")
}

func TestGetBaseDir(t *testing.T) {
	require.Equal(t, configBaseDir, GetBaseDir(), "Incorrect base directory")
}

func TestGetCassandraFileName(t *testing.T) {
	require.Equal(t, cassandraConfigFileName, GetCassandraFileName(), "Incorrect Cassandra filename")
}

func TestGetLoggerFileName(t *testing.T) {
	require.Equal(t, loggerConfigFileName, GetLoggerFileName(), "Incorrect logger filename")
}

func TestGetIntegrationTestKeyspaceSuffix(t *testing.T) {
	require.Equal(t, integrationTestKeyspaceSuffix, GetIntegrationTestKeyspaceSuffix(), "Incorrect integration test keyspace suffix")
}

func TestGetCassandraPrefix(t *testing.T) {
	require.Equal(t, cassandraPrefix, GetCassandraPrefix(), "Incorrect Cassandra environment prefix")
}

func TestGetLoggerPrefix(t *testing.T) {
	require.Equal(t, loggerPrefix, GetLoggerPrefix(), "Incorrect Zap logger environment prefix")
}

func TestGetCassandraMaxConnectRetries(t *testing.T) {
	require.Equal(t, cassandraMaxConnectRetries, GetCassandraMaxConnectRetries(), "Incorrect Cassandra connection retries")
}

func TestGetGithubCIKey(t *testing.T) {
	require.Equal(t, githubCIKey, GetGithubCIKey(), "Incorrect Github CI environment key")
}

func TestGetAuthFileName(t *testing.T) {
	require.Equal(t, authConfigFileName, GetAuthFileName(), "Incorrect authentication filename")
}

func TestGetAuthPrefix(t *testing.T) {
	require.Equal(t, authPrefix, GetAuthPrefix(), "Incorrect authorization environment prefix")
}

func TestGetHTTPRESTFileName(t *testing.T) {
	require.Equal(t, restConfigFileName, GetHTTPRESTFileName(), "Incorrect HTTP REST filename")
}

func TestGetHTTPRESTPrefix(t *testing.T) {
	require.Equal(t, restPrefix, GetHTTPRESTPrefix(), "Incorrect HTTP REST environment prefix")
}

func TestGetGraphQLFileName(t *testing.T) {
	require.Equal(t, graphqlConfigFileName, GetGraphQLFileName(), "Incorrect HTTP GraphQL filename")
}

func TestGetGraphQLPrefix(t *testing.T) {
	require.Equal(t, graphqlPrefix, GetGraphQLPrefix(), "Incorrect HTTP GraphQL environment prefix")
}

func TestGetDeleteUserAccountConfirmation(t *testing.T) {
	require.Equal(t, deleteUserAccountConfirmation, GetDeleteUserAccountConfirmation(), "Incorrect user account deletion confirmation message.")
}

func TestGetRedisFileName(t *testing.T) {
	require.Equal(t, redisConfigFileName, GetRedisFileName(), "Incorrect Redis filename")
}

func TestGetRedisPrefix(t *testing.T) {
	require.Equal(t, redisPrefix, GetRedisPrefix(), "Incorrect Redis environment prefix")
}

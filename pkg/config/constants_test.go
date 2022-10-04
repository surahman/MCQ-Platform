package config

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

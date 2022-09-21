package config

const (
	// Configuration file directories
	configEtcDir  = "/etc/CassandraTutorial.conf/"
	configHomeDir = "$HOME/.CassandraTutorial/"

	// Configuration file names
	cassandraConfigFileName = "CassandraConfig.yaml"

	// Environment variables
	cassandraPrefix = "CASSANDRA"
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

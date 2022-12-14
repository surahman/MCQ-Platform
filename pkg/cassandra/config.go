package cassandra

import (
	"github.com/spf13/afero"
	"github.com/surahman/mcq-platform/pkg/config_loader"
	"github.com/surahman/mcq-platform/pkg/constants"
)

// config is the configuration container for connecting to the Cassandra cluster
type config struct {
	Authentication struct {
		Username string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username" validate:"required"`
		Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password" validate:"required"`
	} `json:"authentication,omitempty" yaml:"authentication,omitempty" mapstructure:"authentication"`
	Keyspace struct {
		Name              string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name" validate:"required"`
		ReplicationClass  string `json:"replication_class,omitempty" yaml:"replication_class,omitempty" mapstructure:"replication_class" validate:"required"`
		ReplicationFactor int    `json:"replication_factor,omitempty" yaml:"replication_factor,omitempty" mapstructure:"replication_factor" validate:"required,numeric,min=1"`
	} `json:"keyspace,omitempty" yaml:"keyspace,omitempty" mapstructure:"keyspace"`
	Connection struct {
		Consistency     string   `json:"consistency,omitempty" yaml:"consistency,omitempty" mapstructure:"consistency" validate:"required"`
		ClusterIP       []string `json:"cluster_ip,omitempty" yaml:"cluster_ip,omitempty" mapstructure:"cluster_ip" validate:"required,min=1"`
		ProtoVersion    int      `json:"proto_version,omitempty" yaml:"proto_version,omitempty" mapstructure:"proto_version" validate:"required,numeric,min=4"`
		MaxConnAttempts int      `json:"max_connection_attempts,omitempty" yaml:"max_connection_attempts,omitempty" mapstructure:"max_connection_attempts" validate:"required,min=1"`
		Timeout         int      `json:"timeout,omitempty" yaml:"timeout,omitempty" mapstructure:"timeout" validate:"required,numeric,min=1"`
	} `json:"connection,omitempty" yaml:"connection,omitempty" mapstructure:"connection"`
}

// newConfig creates a blank configuration struct for Cassandra.
func newConfig() *config {
	return &config{}
}

// Load will attempt to load configurations from a file on a file system and then overwrite values using environment variables.
func (cfg *config) Load(fs afero.Fs) (err error) {
	return config_loader.ConfigLoader(fs, cfg, constants.GetCassandraFileName(), constants.GetCassandraPrefix(), "yaml")
}

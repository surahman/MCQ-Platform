package config

import (
	"errors"

	"github.com/spf13/afero"
)

// Type is an enum for the type of configuration file to be created.
type Type uint16

// Enum values indicating the configuration types.
const (
	Cassandra = iota
	Redis
	Authorization
)

// IConfig is the base configuration type interface.
type IConfig interface {
	Load(afero.Fs) error // Loads the configuration file from a supplied file system.
}

// Factory will return a blank config struct to be populated.
func Factory(configType Type) (IConfig, error) {
	switch configType {
	case Cassandra:
		return newCassandraConfig(), nil
	default:
		return nil, errors.New("invalid config type provided")
	}
}

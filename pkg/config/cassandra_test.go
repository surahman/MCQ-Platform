package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCassandraConfig(t *testing.T) {
	conf := newCassandraConfig()
	require.NotNilf(t, conf, "Should return a non-nil config struct")
	require.True(t, reflect.TypeOf(conf) == reflect.TypeOf(&CassandraConfig{}), "Should return a CassandraConfig")
}

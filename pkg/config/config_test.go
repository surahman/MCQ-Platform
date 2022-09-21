package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigFactory(t *testing.T) {

	testCases := []struct {
		name          string
		option        Type
		expectedType  reflect.Type
		expectedError require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"Invalid config type",
			99,
			nil,
			require.Error,
		}, {
			"Cassandra config",
			Cassandra,
			reflect.TypeOf(&CassandraConfig{}),
			require.NoError,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualConfig, err := Factory(testCase.option)
			testCase.expectedError(t, err)
			require.Equal(t, testCase.expectedType, reflect.TypeOf(actualConfig))
		})
	}
}

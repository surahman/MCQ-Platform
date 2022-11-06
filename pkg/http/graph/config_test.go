package graphql

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/validator"
	"gopkg.in/yaml.v3"
)

func TestRestConfigs_Load(t *testing.T) {
	keyspaceGen := constants.GetGraphQLPrefix() + "_SERVER."
	keyspaceAuth := constants.GetGraphQLPrefix() + "_AUTHORIZATION."

	testCases := []struct {
		name      string
		input     string
		expectErr require.ErrorAssertionFunc
		expectLen int
	}{
		// ----- test cases start ----- //
		{
			"empty - etc dir",
			graphqlConfigTestData["empty"],
			require.Error,
			5,
		}, {
			"valid - etc dir",
			graphqlConfigTestData["valid"],
			require.NoError,
			0,
		}, {
			"out of range port - etc dir",
			graphqlConfigTestData["out of range port"],
			require.Error,
			1,
		}, {
			"no base path - etc dir",
			graphqlConfigTestData["no base path"],
			require.Error,
			1,
		}, {
			"no swagger path - etc dir",
			graphqlConfigTestData["no swagger path"],
			require.Error,
			1,
		}, {
			"no auth header - etc dir",
			graphqlConfigTestData["no auth header"],
			require.Error,
			1,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetGraphQLFileName(), []byte(testCase.input), 0644), "Failed to write in memory file")

			// Load from mock filesystem.
			actual := &config{}
			err := actual.Load(fs)
			testCase.expectErr(t, err)

			if err != nil {
				require.Equalf(t, testCase.expectLen, len(err.(*validator.ErrorValidation).Errors), "Expected errors count is incorrect: %v", err)
				return
			}

			// Load expected struct.
			expected := &config{}
			require.NoError(t, yaml.Unmarshal([]byte(testCase.input), expected), "failed to unmarshal expected constants")
			require.True(t, reflect.DeepEqual(expected, actual))

			// Test configuring of environment variable.
			basePath := xid.New().String()
			playgroundPath := xid.New().String()
			headerKey := xid.New().String()
			portNumber := 1600
			shutdownDelay := 36
			t.Setenv(keyspaceGen+"BASE_PATH", basePath)
			t.Setenv(keyspaceGen+"PLAYGROUND_PATH", playgroundPath)
			t.Setenv(keyspaceGen+"PORT_NUMBER", strconv.Itoa(portNumber))
			t.Setenv(keyspaceGen+"SHUTDOWN_DELAY", strconv.Itoa(shutdownDelay))
			t.Setenv(keyspaceAuth+"HEADER_KEY", headerKey)
			err = actual.Load(fs)
			require.NoErrorf(t, err, "Failed to load constants file: %v", err)
			require.Equal(t, basePath, actual.Server.BasePath, "Failed to load base path environment variable into configs")
			require.Equal(t, playgroundPath, actual.Server.PlaygroundPath, "Failed to load playground path environment variable into configs")
			require.Equal(t, portNumber, actual.Server.PortNumber, "Failed to load port environment variable into configs")
			require.Equal(t, shutdownDelay, actual.Server.ShutdownDelay, "Failed to load shutdown delay environment variable into configs")
			require.Equal(t, headerKey, actual.Authorization.HeaderKey, "Failed to load authorization header key environment variable into configs")
		})
	}
}

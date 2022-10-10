package rest

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
	keyspaceGen := constants.GetHTTPRESTPrefix() + "_GENERAL."

	testCases := []struct {
		name      string
		input     string
		expectErr require.ErrorAssertionFunc
		expectLen int
	}{
		// ----- test cases start ----- //
		{
			"empty - etc dir",
			restConfigTestData["empty"],
			require.Error,
			4,
		}, {
			"valid - etc dir",
			restConfigTestData["valid"],
			require.NoError,
			0,
		}, {
			"out of range port - etc dir",
			restConfigTestData["out of range port"],
			require.Error,
			1,
		}, {
			"no base path- etc dir",
			restConfigTestData["no base path"],
			require.Error,
			1,
		}, {
			"no swagger path- etc dir",
			restConfigTestData["no swagger path"],
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
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetHTTPRESTFileName(), []byte(testCase.input), 0644), "Failed to write in memory file")

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
			swaggerPath := xid.New().String()
			portNumber := 1600
			shutdownDelay := 36
			t.Setenv(keyspaceGen+"BASE_PATH", basePath)
			t.Setenv(keyspaceGen+"SWAGGER_PATH", swaggerPath)
			t.Setenv(keyspaceGen+"PORT_NUMBER", strconv.Itoa(portNumber))
			t.Setenv(keyspaceGen+"SHUTDOWN_DELAY", strconv.Itoa(shutdownDelay))
			err = actual.Load(fs)
			require.NoErrorf(t, err, "Failed to load constants file: %v", err)
			require.Equal(t, basePath, actual.General.BasePath, "Failed to load base path environment variable into configs")
			require.Equal(t, swaggerPath, actual.General.SwaggerPath, "Failed to load swagger path environment variable into configs")
			require.Equal(t, portNumber, actual.General.PortNumber, "Failed to load port environment variable into configs")
			require.Equal(t, shutdownDelay, actual.General.ShutdownDelay, "Failed to load shutdown delay environment variable into configs")
		})
	}
}

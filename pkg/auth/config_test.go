package auth

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/rs/xid"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/constants"
	"gopkg.in/yaml.v3"
)

func TestAuthConfigs_Load(t *testing.T) {
	keyspaceJwt := constants.GetAuthPrefix() + "_JWT."
	keyspaceGen := constants.GetAuthPrefix() + "_GENERAL."

	testCases := []struct {
		name      string
		input     string
		expectErr require.ErrorAssertionFunc
		expectLen int
	}{
		// ----- test cases start ----- //
		{
			"empty - etc dir",
			authConfigTestData["empty"],
			require.Error,
			3,
		},
		{
			"valid - etc dir",
			authConfigTestData["valid"],
			require.NoError,
			0,
		},
		{
			"no issuer - etc dir",
			authConfigTestData["no_issuer"],
			require.Error,
			1,
		},
		{
			"bcrypt cost below 4 - etc dir",
			authConfigTestData["bcrypt_cost_below_4"],
			require.Error,
			1,
		},
		{
			"bcrypt cost above 31 - etc dir",
			authConfigTestData["bcrypt_cost_above_31"],
			require.Error,
			1,
		},
		{
			"jwt expiration below 10s - etc dir",
			authConfigTestData["jwt_expiration_below_60s"],
			require.Error,
			1,
		},
		{
			"jwt key below 8 - etc dir",
			authConfigTestData["jwt_key_below_8"],
			require.Error,
			1,
		},
		{
			"jwt key above 256 - etc dir",
			authConfigTestData["jwt_key_above_256"],
			require.Error,
			1,
		},
		{
			"low refresh threshold - etc dir",
			authConfigTestData["low_refresh_threshold"],
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
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetAuthFileName(), []byte(testCase.input), 0644), "Failed to write in memory file")

			// Load from mock filesystem.
			actual := &config{}
			err := actual.Load(fs)
			testCase.expectErr(t, err)

			if err != nil {
				return
			}

			// Load expected struct.
			expected := &config{}
			require.NoError(t, yaml.Unmarshal([]byte(testCase.input), expected), "failed to unmarshal expected constants")
			require.True(t, reflect.DeepEqual(expected, actual))

			// Test configuring of environment variable.
			testKey := xid.New().String()
			testExpDur := int64(999)
			testRefThreshold := int64(555)
			testBcryptCost := 16
			testIssuer := "test issuer"
			t.Setenv(keyspaceJwt+"KEY", testKey)
			t.Setenv(keyspaceJwt+"ISSUER", testIssuer)
			t.Setenv(keyspaceJwt+"EXPIRATION_DURATION", strconv.FormatInt(testExpDur, 10))
			t.Setenv(keyspaceJwt+"REFRESH_THRESHOLD", strconv.FormatInt(testRefThreshold, 10))
			t.Setenv(keyspaceGen+"BCRYPT_COST", strconv.Itoa(testBcryptCost))
			err = actual.Load(fs)
			require.NoErrorf(t, err, "Failed to load constants file: %v", err)
			require.Equalf(t, testKey, actual.JWTConfig.Key, "Failed to load key environment variable into configs")
			require.Equalf(t, testIssuer, actual.JWTConfig.Issuer, "Failed to load issuer environment variable into configs")
			require.Equalf(t, testExpDur, actual.JWTConfig.ExpirationDuration, "Failed to load duration environment variable into configs")
			require.Equalf(t, testRefThreshold, actual.JWTConfig.RefreshThreshold, "Failed to load refresh threshold environment variable into configs")
			require.Equalf(t, testBcryptCost, actual.General.BcryptCost, "Failed to load bcrypt cost environment variable into configs")
		})
	}
}

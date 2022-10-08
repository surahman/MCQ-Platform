package auth

import (
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/logger"
	"github.com/surahman/mcq-platform/pkg/model/http"
)

func TestNewAuth(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
	require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+constants.GetAuthFileName(),
		[]byte(authConfigTestData["valid"]), 0644), "Failed to write in memory file")

	testCases := []struct {
		name      string
		fs        *afero.Fs
		log       *logger.Logger
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"Invalid file system and logger",
			nil,
			nil,
			require.Error,
			require.Nil,
		}, {
			"Invalid file system",
			nil,
			zapLogger,
			require.Error,
			require.Nil,
		}, {
			"Invalid logger",
			&fs,
			nil,
			require.Error,
			require.Nil,
		}, {
			"Valid",
			&fs,
			zapLogger,
			require.NoError,
			require.NotNil,
		},
		// ----- test cases end ----- //
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cassandra, err := NewAuth(testCase.fs, testCase.log)
			testCase.expectErr(t, err)
			testCase.expectNil(t, cassandra)
		})
	}
}

func TestNewAuthImpl(t *testing.T) {
	testCases := []struct {
		name      string
		fileName  string
		input     string
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"File found",
			constants.GetAuthFileName(),
			authConfigTestData["valid"],
			require.NoError,
			require.NotNil,
		}, {
			"File not found",
			"wrong_file_name.yaml",
			authConfigTestData["valid"],
			require.Error,
			require.Nil,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Configure mock filesystem.
			fs := afero.NewMemMapFs()
			require.NoError(t, fs.MkdirAll(constants.GetEtcDir(), 0644), "Failed to create in memory directory")
			require.NoError(t, afero.WriteFile(fs, constants.GetEtcDir()+testCase.fileName, []byte(testCase.input), 0644), "Failed to write in memory file")

			c, err := NewAuth(&fs, zapLogger)
			testCase.expectErr(t, err)
			testCase.expectNil(t, c)
		})
	}
}

func TestAuthImpl_HashPassword(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		expectErr require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"Empty password",
			"",
			require.NoError,
		}, {
			"Valid",
			"ELy@FRrn7DW8Cj1QQj^zG&X%$9cjVU4R",
			require.NoError,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := testAuth.HashPassword(testCase.input)
			testCase.expectErr(t, err)
		})
	}
}

func TestAuthImpl_CheckPassword(t *testing.T) {
	testCases := []struct {
		name      string
		plaintext string
		hashed    string
		expectErr require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"Empty password's hash",
			"",
			"$2a$08$vZhD311uyi8FnkjyoT.1req7ixf0CXRARozPTaj4gnhr/F3m/q7NW",
			require.NoError,
		}, {
			"Valid password's hash",
			"ELy@FRrn7DW8Cj1QQj^zG&X%$9cjVU4R",
			"$2a$08$YXYc8lyxnS7VPy6f28Gmd.udRTVrxKewXX9E3ULs0/ynkTL6PY/0K",
			require.NoError,
		}, {
			"Invalid password's hash",
			"$2a$08$YXYc8lyxnS7VPy6f28Gmd.udRTVrxKewXX9E3ULs0/ynkTL6PY/0K",
			"$2a$08$YXYc8lyxnS7VPy6f28Gmd.udRTVrxKewXX9E3ULs0/ynkTL6PY/0K",
			require.Error,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testAuth.CheckPassword(testCase.hashed, testCase.plaintext)
			testCase.expectErr(t, err)
		})
	}
}

func TestAuthImpl_GenerateJWT(t *testing.T) {
	userName := "test username"
	authResponse, err := testAuth.GenerateJWT(userName)
	require.NoError(t, err, "JWT creation failed")
	require.True(t, authResponse.Expires.After(time.Now()), "JWT expires before current time")
	require.True(t, authResponse.Expires.Before(time.Now().Add(time.Duration(expirationDuration+1)*time.Second)), "JWT expires after deadline")

	// Check for username in claim.
	actualUname, err := testAuth.UsernameFromJWT(authResponse.Token)
	require.NoError(t, err, "failed to extract username from JWT")
	require.Equalf(t, userName, actualUname, "incorrect username retrieved from JWT")

	// Validate token.
	require.NoError(t, testAuth.ValidateJWT(authResponse.Token), "failed to validate a valid token")
}

func TestAuthImpl_ValidateJWT(t *testing.T) {
	var err error
	var testAuthImpl *authImpl
	var testJWT *model_http.JWTAuthResponse

	// Testing for an invalid issuer.
	testAuthImpl, err = getTestConfiguration()
	require.NoError(t, err, "failed to generate test authorization for issuer")
	testAuthImpl.conf.JWTConfig.Issuer = "an invalid issuer"
	testJWT, err = testAuthImpl.GenerateJWT("test username")
	require.NoError(t, err, "failed to create invalid issuer JWT")

	err = testAuth.ValidateJWT(testJWT.Token)
	require.Error(t, err, "validation of invalid issuer token should fail")
	require.Contains(t, err.Error(), "issuer", "error should be about an invalid issuer")

	// Test for expiration.
	testAuthImpl, err = getTestConfiguration()
	require.NoError(t, err, "failed to generate test authorization for expiration")
	testAuthImpl.conf.JWTConfig.ExpirationDuration = 1
	testJWT, err = testAuthImpl.GenerateJWT("test username")
	require.NoError(t, err, "failed to create expiration JWT")

	time.Sleep(time.Duration(testAuthImpl.conf.JWTConfig.ExpirationDuration+1) * time.Second)
	err = testAuthImpl.ValidateJWT(testJWT.Token)
	require.Error(t, err, "validation of expired token should fail")
	require.Contains(t, err.Error(), "expired", "error should be about expiration")

	// Test for claim parsing.
	testAuthImpl, err = getTestConfiguration()
	require.NoError(t, err, "failed to generate test authorization for claim parsing")

	err = testAuthImpl.ValidateJWT("")
	require.Error(t, err, "parsing an empty token should fail")

	err = testAuthImpl.ValidateJWT("bad#token#string")
	require.Error(t, err, "parsing and invalid token should fail")
}

func TestAuthImpl_UsernameFromJWT(t *testing.T) {
	var err error

	// Empty token.
	_, err = testAuth.UsernameFromJWT("")
	require.Error(t, err, "empty token sting should fail")

	// Valid token.
	username := "test username"
	token, err := testAuth.GenerateJWT(username)
	require.NoError(t, err, "failed to generate test token")
	var actualUsername string
	actualUsername, err = testAuth.UsernameFromJWT(token.Token)
	require.NoError(t, err, "failed to extract username from valid token")
	require.Equal(t, username, actualUsername, "expected and actual username do not match")

}

func TestAuthImpl_RefreshJWT(t *testing.T) {

	testCases := []struct {
		name               string
		testUsername       string
		expirationDuration int64
		sleepTime          int
		expectErr          require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"Valid token",
			"test username",
			4,
			2,
			require.NoError,
		}, {
			"Invalid token",
			"test username",
			1,
			2,
			require.Error,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var err error
			var testAuthImpl *authImpl
			var testJWT *model_http.JWTAuthResponse
			var refreshedToken *model_http.JWTAuthResponse
			var actualUsername string

			// Test for valid token refresh.
			testAuthImpl, err = getTestConfiguration()
			testAuthImpl.conf.JWTConfig.ExpirationDuration = testCase.expirationDuration
			require.NoError(t, err, "failed to generate test authorization")
			testJWT, err = testAuthImpl.GenerateJWT(testCase.testUsername)
			require.NoError(t, err, "failed to create initial JWT")
			require.NoError(t, testAuthImpl.ValidateJWT(testJWT.Token))

			time.Sleep(time.Duration(testCase.sleepTime) * time.Second)
			refreshedToken, err = testAuthImpl.RefreshJWT(testJWT.Token)
			testCase.expectErr(t, err, "error case when refreshing JWT failed")

			if err != nil {
				return
			}

			require.True(t,
				refreshedToken.Expires.After(time.Now().Add(time.Duration(testAuthImpl.conf.JWTConfig.ExpirationDuration-1)*time.Second)),
				"token expires before the required deadline")

			actualUsername, err = testAuthImpl.UsernameFromJWT(testJWT.Token)
			require.NoErrorf(t, err, "failed to extract username from JWT")
			require.Equal(t, testCase.testUsername, actualUsername, "failed to extract correct username from JWT")
		})
	}
}

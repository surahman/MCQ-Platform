package auth

import (
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/constants"
	"github.com/surahman/mcq-platform/pkg/logger"
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

	actualUname, err := testAuth.UsernameFromJWT(authResponse.Token)
	require.NoError(t, err, "failed to extract username from JWT")
	require.Equalf(t, userName, actualUname, "incorrect username retrieved from JWT")
}

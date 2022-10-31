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
	require.True(t, authResponse.Expires > time.Now().Unix(), "JWT expires before current time")
	require.True(t, authResponse.Expires < time.Now().Add(time.Duration(expirationDuration+1)*time.Second).Unix(), "JWT expires after deadline")

	// Check validate token and check for username in claim.
	actualUname, expiresAt, err := testAuth.ValidateJWT(authResponse.Token)
	require.NoError(t, err, "failed to extract username from JWT")
	require.Equalf(t, userName, actualUname, "incorrect username retrieved from JWT")
	require.True(t, expiresAt > 0, "invalid expiration time")
}

func TestAuthImpl_ValidateJWT(t *testing.T) {
	t.Run("JWT claim parsing tests", func(t *testing.T) {
		testAuthImpl, err := getTestConfiguration()
		require.NoError(t, err, "failed to generate test authorization for claim parsing")

		_, _, err = testAuthImpl.ValidateJWT("")
		require.Error(t, err, "parsing an empty token should fail")

		_, _, err = testAuthImpl.ValidateJWT("bad#token#string")
		require.Error(t, err, "parsing and invalid token should fail")
	})

	const testUsername = "test username"
	testCases := []struct {
		name               string
		issuerName         string
		expectErrMsg       string
		expirationDuration int64
		expectErr          require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"Valid token",
			"",
			"",
			0,
			require.NoError,
		}, {
			"Invalid issuer token",
			"some random name",
			"issuer",
			0,
			require.Error,
		}, {
			"Invalid expired token",
			"some random name",
			"expired",
			1,
			require.Error,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Test authorization config and token generation.
			testAuthImpl, err := getTestConfiguration()
			require.NoError(t, err, "failed to generate test authorization for issuer")
			if testCase.issuerName != "" {
				testAuthImpl.conf.JWTConfig.Issuer = testCase.issuerName
			}
			if testCase.expirationDuration != 0 {
				testAuthImpl.conf.JWTConfig.ExpirationDuration = testCase.expirationDuration
			}

			// Generate test token.
			testJWT, err := testAuthImpl.GenerateJWT(testUsername)
			require.NoError(t, err, "failed to create test JWT")

			// Conditional sleep to expire token.
			if testCase.expirationDuration > 0 {
				time.Sleep(time.Duration(testCase.expirationDuration+1) * time.Second)
			}

			username, expiresAt, err := testAuth.ValidateJWT(testJWT.Token)
			testCase.expectErr(t, err, "validation of issued token error condition failed")

			if err != nil {
				require.Contains(t, err.Error(), testCase.expectErrMsg, "error message did not contain expected err")
				return
			}
			require.Equal(t, testUsername, username, "username failed to match the expected")
			require.True(t, expiresAt > 0, "invalid expiration time")
		})
	}
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
			// Test authorization config and token generation.
			testAuthImpl, err := getTestConfiguration()
			testAuthImpl.conf.JWTConfig.ExpirationDuration = testCase.expirationDuration
			require.NoError(t, err, "failed to generate test authorization")
			testJWT, err := testAuthImpl.GenerateJWT(testCase.testUsername)
			require.NoError(t, err, "failed to create initial JWT")
			actualUsername, expiresAt, err := testAuthImpl.ValidateJWT(testJWT.Token)
			require.NoError(t, err, "failed to validate original test token")
			require.Equal(t, testCase.testUsername, actualUsername, "failed to extract correct username from original JWT")
			require.True(t, expiresAt > 0, "invalid expiration time of original token")

			time.Sleep(time.Duration(testCase.sleepTime) * time.Second)
			refreshedToken, err := testAuthImpl.RefreshJWT(testJWT.Token)
			testCase.expectErr(t, err, "error case when refreshing JWT failed")

			if err != nil {
				return
			}

			require.True(t,
				refreshedToken.Expires > time.Now().Add(time.Duration(testAuthImpl.conf.JWTConfig.ExpirationDuration-1)*time.Second).Unix(),
				"token expires before the required deadline")

			actualUsername, expiresAt, err = testAuthImpl.ValidateJWT(testJWT.Token)
			require.NoErrorf(t, err, "failed to validate refreshed JWT")
			require.Equal(t, testCase.testUsername, actualUsername, "failed to extract correct username from refreshed JWT")
			require.True(t, expiresAt > 0, "invalid expiration time of refreshed token")
		})
	}
}

func TestAuthImpl_RefreshThreshold(t *testing.T) {
	require.Equal(t, refreshThreshold, testAuth.RefreshThreshold(), "token refresh threshold did not match expected threshold")
}

func TestAuthImpl_encryptAES256_and_decryptAES256(t *testing.T) {
	testCases := []struct {
		name      string
		plainText string
		expectStr require.BoolAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name:      "to string",
			plainText: "this text should be encrypted",
			expectStr: require.True,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			// Encrypt phase.
			cipherStr, cipherBytes, err := testAuth.encryptAES256([]byte(testCase.plainText))
			require.NoError(t, err, "error encrypting to AES256")
			require.NotNil(t, cipherBytes, "no cipher block returned as bytes")
			testCase.expectStr(t, len(cipherStr) > 0, "cipher string expectation failed")

			// Decrypt phase.
			plainTextBytes, err := testAuth.decryptAES256(cipherBytes)
			require.NoError(t, err, "error decrypting from AES256")
			require.NotNil(t, plainTextBytes, "no plaintext block returned as bytes")
			require.Equal(t, testCase.plainText, string(plainTextBytes), "decrypted cipher does not match input plaintext")
		})
	}
}

func TestAuthImpl_Encrypt_Decrypt_String(t *testing.T) {
	toEncrypt := "this is a text string to be encrypted/decrypted"

	ciphertext, err := testAuth.EncryptToString([]byte(toEncrypt))
	require.NoError(t, err, "encrypt to string failed")
	require.True(t, len(ciphertext) > 0, "encrypted string is empty")

	plaintext, err := testAuth.DecryptFromString(ciphertext)
	require.NoError(t, err, "encrypt from string failed")
	require.Equal(t, toEncrypt, string(plaintext), "decrypted string does not match original")
}

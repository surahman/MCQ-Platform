package model_cassandra

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/validator"
)

func TestValidateUser(t *testing.T) {
	testCases := []struct {
		name        string
		input       *UserAccount
		expectErr   require.ErrorAssertionFunc
		expectedLen int
	}{
		// ----- test cases start ----- //
		{
			name:        "Empty",
			input:       &UserAccount{},
			expectErr:   require.Error,
			expectedLen: 5,
		}, {
			name: "Valid",
			input: &UserAccount{
				Username:  "username1",
				Password:  "password-1",
				FirstName: "first name",
				LastName:  "last name",
				Email:     "username@email.com",
			},
			expectErr:   require.NoError,
			expectedLen: 0,
		}, {
			name: "Invalid username",
			input: &UserAccount{
				Username:  "username-1",
				Password:  "password-1",
				FirstName: "first name",
				LastName:  "last name",
				Email:     "username@email.com",
			},
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name: "Invalid password",
			input: &UserAccount{
				Username:  "username1",
				Password:  "password-1password-1password-1password-1",
				FirstName: "first name",
				LastName:  "last name",
				Email:     "username@email.com",
			},
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name: "Invalid email",
			input: &UserAccount{
				Username:  "username1",
				Password:  "password-1",
				FirstName: "first name",
				LastName:  "last name",
				Email:     "username@email",
			},
			expectErr:   require.Error,
			expectedLen: 1,
		},
		// ----- test cases end ----- //
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validator.ValidateStruct(testCase.input)
			testCase.expectErr(t, err)

			if err != nil {
				require.Equal(t, testCase.expectedLen, len(err.(*validator.ErrorValidation).Errors))
			}
		})
	}
}

func TestBlake2b256(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		// ----- test cases start ----- //
		{
			name:     "Generic",
			input:    "account-id-1",
			expected: "mAHpa8iePo3zmyxx_kMWulKeiRtP-KIm-Kq2qr4vKdM=",
		}, {
			name:     "Random 8 character",
			input:    "K40c&9*H",
			expected: "yKaKVtbY28qtEnBil1lsglC3Rw3HKd_K9Ex2hasqlAc=",
		}, {
			name:     "Random 32 character",
			input:    "yyx86kaBXF9bUn2w1I6m5efNMs&rOjZd",
			expected: "MN_76A0UaucyVNqw9M8IKs6yQg4UFl_EfKPDCflTigg=",
		},
		// ----- test cases end ----- //
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := Blake2b256(testCase.input)
			require.Equal(t, testCase.expected, actual)
		})
	}
}

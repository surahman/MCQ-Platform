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
				UserLoginCredentials: UserLoginCredentials{
					Username: "username1",
					Password: "password-1",
				},
				FirstName: "first name",
				LastName:  "last name",
				Email:     "username@email.com",
			},
			expectErr:   require.NoError,
			expectedLen: 0,
		}, {
			name: "Invalid username",
			input: &UserAccount{
				UserLoginCredentials: UserLoginCredentials{
					Username: "username 1",
					Password: "password-1",
				},
				FirstName: "first name",
				LastName:  "last name",
				Email:     "username@email.com",
			},
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name: "Invalid password",
			input: &UserAccount{
				UserLoginCredentials: UserLoginCredentials{
					Username: "username1",
					Password: "password-1password-1password-1password-1",
				},
				FirstName: "first name",
				LastName:  "last name",
				Email:     "username@email.com",
			},
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name: "Invalid email",
			input: &UserAccount{
				UserLoginCredentials: UserLoginCredentials{
					Username: "username1",
					Password: "password-1",
				},
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

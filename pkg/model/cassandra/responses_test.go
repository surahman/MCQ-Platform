package model_cassandra

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/validator"
)

func TestValidateQuizResponse(t *testing.T) {
	testCases := []struct {
		name        string
		input       *QuizResponse
		expectErr   require.ErrorAssertionFunc
		expectedLen int
	}{
		// ----- test cases start ----- //
		{
			name:        "Valid response with one question",
			input:       &QuizResponse{[][]int32{{0}}},
			expectErr:   require.NoError,
			expectedLen: 0,
		}, {
			name:        "Valid response with ten question",
			input:       &QuizResponse{[][]int32{{0}, {1}, {2}, {2}, {2}, {2}, {2}, {2}, {2}, {2}}},
			expectErr:   require.NoError,
			expectedLen: 0,
		}, {
			name:        "Invalid response with too many question",
			input:       &QuizResponse{[][]int32{{0}, {1}, {2}, {2}, {2}, {2}, {2}, {2}, {2}, {2}, {2}}},
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name:        "Invalid response with too many answers",
			input:       &QuizResponse{[][]int32{{0}, {1}, {0, 1, 2, 3, 4, 4}, {2}, {2}, {2}, {2}, {2}, {2}, {2}}},
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name:        "Invalid response with +ve out of range answers",
			input:       &QuizResponse{[][]int32{{0}, {1}, {0, 1, 2, 3, 5}, {2}, {2}, {2}, {2}, {2}, {2}, {2}}},
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name:        "Invalid response with -ve out of range answers",
			input:       &QuizResponse{[][]int32{{0}, {1}, {0, -1, 2, 3}, {2}, {2}, {2}, {2}, {2}, {2}, {2}}},
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name:        "Valid response with no answers",
			input:       &QuizResponse{[][]int32{{}}},
			expectErr:   require.NoError,
			expectedLen: 0,
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

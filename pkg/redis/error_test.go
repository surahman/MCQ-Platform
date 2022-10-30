package redis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	testCases := []struct {
		name         string
		err          *Error
		expectedCode int
		expectedType any
	}{
		// ----- test cases start ----- //
		{
			name:         "base error",
			err:          NewError("base error"),
			expectedCode: ErrorUnknown,
		}, {
			name:         "cache miss",
			err:          NewError("cache miss").errorCacheMiss(),
			expectedCode: ErrorCacheMiss,
		}, {
			name:         "cache set",
			err:          NewError("cache set").errorCacheSet(),
			expectedCode: ErrorCacheSet,
		}, {
			name:         "cache del",
			err:          NewError("cache del").errorCacheDel(),
			expectedCode: ErrorCacheDel,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.NotNil(t, testCase.err, "error should not be nil")
			require.Equal(t, testCase.expectedCode, testCase.err.Code, "expected error code did not match")
			require.Equal(t, testCase.name, testCase.err.Message, "error messages did not match")
		})
	}
}

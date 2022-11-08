package graphql_resolvers

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGinContextFromContext(t *testing.T) {
	testCases := []struct {
		name        string
		expectedMsg string
		expectErr   require.ErrorAssertionFunc
		ctx         context.Context
	}{
		// ----- test cases start ----- //
		{
			name:        "no context",
			expectedMsg: "information not found",
			expectErr:   require.Error,
			ctx:         context.TODO(),
		}, {
			name:        "incorrect context",
			expectedMsg: "information malformed",
			expectErr:   require.Error,
			ctx:         context.WithValue(context.TODO(), "GinContextKey", context.TODO()),
		}, {
			name:      "success",
			expectErr: require.NoError,
			ctx:       context.WithValue(context.TODO(), "GinContextKey", &gin.Context{}),
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			_, err := GinContextFromContext(testCase.ctx, zapLogger)

			testCase.expectErr(t, err, "error expectation failed")
			if err != nil {
				require.Contains(t, err.Error(), testCase.expectedMsg, "incorrect error message returned")
			}
		})
	}
}

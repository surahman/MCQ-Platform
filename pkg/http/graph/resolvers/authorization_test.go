package graphql_resolvers

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	http_common "github.com/surahman/mcq-platform/pkg/http"
	"github.com/surahman/mcq-platform/pkg/mocks"
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
			ctx:         context.WithValue(context.TODO(), GinContextKey{}, context.TODO()),
		}, {
			name:      "success",
			expectErr: require.NoError,
			ctx:       context.WithValue(context.TODO(), GinContextKey{}, &gin.Context{}),
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

func TestAuthorizationCheck(t *testing.T) {

	ginCtxNoAuth := &gin.Context{Request: &http.Request{Header: http.Header{}}}
	ginCtxNoAuth.Request.Header.Add(testAuthHeaderKey, "")

	ginCtxAuth := &gin.Context{Request: &http.Request{Header: http.Header{}}}
	ginCtxAuth.Request.Header.Add(testAuthHeaderKey, "test-token")

	testCases := []struct {
		name                string
		expectedMsg         string
		expectErr           require.ErrorAssertionFunc
		ctx                 context.Context
		authValidateJWTData *http_common.MockAuthData
	}{
		// ----- test cases start ----- //
		{
			name:        "no context",
			expectedMsg: "information not found",
			expectErr:   require.Error,
			ctx:         context.TODO(),
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputParam2: int64(-1),
				OutputErr:    nil,
				Times:        0,
			},
		}, {
			name:        "incorrect context",
			expectedMsg: "information malformed",
			expectErr:   require.Error,
			ctx:         context.WithValue(context.TODO(), GinContextKey{}, context.TODO()),
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputParam2: int64(-1),
				OutputErr:    nil,
				Times:        0,
			},
		}, {
			name:        "no token",
			expectedMsg: "does not contain",
			expectErr:   require.Error,
			ctx:         context.WithValue(context.TODO(), GinContextKey{}, ginCtxNoAuth),
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputParam2: int64(-1),
				OutputErr:    nil,
				Times:        0,
			},
		}, {
			name:        "no token",
			expectedMsg: "failed to authenticate token",
			expectErr:   require.Error,
			ctx:         context.WithValue(context.TODO(), GinContextKey{}, ginCtxAuth),
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputParam2: int64(-1),
				OutputErr:    errors.New("failed to authenticate token"),
				Times:        1,
			},
		}, {
			name:      "success",
			expectErr: require.NoError,
			ctx:       context.WithValue(context.TODO(), GinContextKey{}, ginCtxAuth),
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "successful token refresh",
				OutputParam2: int64(999),
				OutputErr:    nil,
				Times:        1,
			},
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.OutputParam1,
				testCase.authValidateJWTData.OutputParam2,
				testCase.authValidateJWTData.OutputErr,
			).Times(testCase.authValidateJWTData.Times)

			username, expiresAt, err := AuthorizationCheck(mockAuth, zapLogger, testAuthHeaderKey, testCase.ctx)

			require.Equal(t, testCase.authValidateJWTData.OutputParam1, username, "expected username does not match")
			require.Equal(t, testCase.authValidateJWTData.OutputParam2, expiresAt, "expected expiration time does not match")

			testCase.expectErr(t, err, "error expectation failed")
			if err != nil {
				require.Contains(t, err.Error(), testCase.expectedMsg, "incorrect error message returned")
			}

		})
	}
}

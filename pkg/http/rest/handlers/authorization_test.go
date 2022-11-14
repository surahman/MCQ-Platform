package http_handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	http_common "github.com/surahman/mcq-platform/pkg/http"
	"github.com/surahman/mcq-platform/pkg/mocks"
)

func TestAuthMiddleware(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockAuth := mocks.NewMockAuth(mockCtrl)

	handler := AuthMiddleware(mockAuth, "Authorization")
	require.NotNil(t, handler)
}

func TestAuthMiddleware_Handler(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		token               string
		expectedStatus      int
		authValidateJWTData *http_common.MockAuthData
	}{
		// ----- test cases start ----- //
		{
			name:           "no token",
			path:           "/no-token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputParam2: int64(-1),
				OutputErr:    nil,
				Times:        0,
			},
		}, {
			name:           "invalid token",
			path:           "/invalid-token",
			token:          "invalid-token",
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputParam2: int64(-1),
				OutputErr:    errors.New("JWT validation failure"),
				Times:        1,
			},
		}, {
			name:           "valid token",
			path:           "/valid-token",
			token:          "valid-token",
			expectedStatus: http.StatusOK,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputParam2: int64(-1),
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

			// Endpoint setup for test.
			router.POST(testCase.path, AuthMiddleware(mockAuth, "Authorization"))
			req, _ := http.NewRequest("POST", testCase.path, nil)
			req.Header.Set("Authorization", testCase.token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}

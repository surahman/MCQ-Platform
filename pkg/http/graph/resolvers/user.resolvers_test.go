package graphql_resolvers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	http_common "github.com/surahman/mcq-platform/pkg/http"
	"github.com/surahman/mcq-platform/pkg/mocks"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
)

func TestMutationResolver_RegisterUser(t *testing.T) {
	router := http_common.GetTestRouter()

	testCases := []struct {
		name                string
		path                string
		user                string
		expectErr           bool
		authHashData        *http_common.MockAuthData
		authGenJWTData      *http_common.MockAuthData
		cassandraCreateData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty user",
			path:      "/register/empty-user",
			user:      fmt.Sprintf(testUserQuery["register"], "", "", "", "", ""),
			expectErr: true,
			authHashData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 0,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
			},
		}, {
			name: "valid user",
			path: "/register/valid-user",
			user: fmt.Sprintf(testUserQuery["register"], "first name", "last name", "email@address.com", "username999", "password999"),
			authHashData: &http_common.MockAuthData{
				OutputParam1: "hashed password",
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 1,
			},
			authGenJWTData: &http_common.MockAuthData{
				OutputParam1: &model_http.JWTAuthResponse{
					Token:     "Test Token",
					Expires:   99,
					Threshold: 100,
				},
				Times: 1,
			},
		}, {
			name:      "password hash failure",
			path:      "/register/pwd-hash-failure",
			user:      fmt.Sprintf(testUserQuery["register"], "first name", "last name", "email@address.com", "username999", "password999"),
			expectErr: true,
			authHashData: &http_common.MockAuthData{
				OutputParam1: "hashed password",
				OutputErr:    errors.New("password hash failure"),
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 0,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
			},
		}, {
			name:      "database failure",
			path:      "/register/database-failure",
			user:      fmt.Sprintf(testUserQuery["register"], "first name", "last name", "email@address.com", "username999", "password999"),
			expectErr: true,
			authHashData: &http_common.MockAuthData{
				OutputParam1: "hashed password",
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{Status: http.StatusNotFound},
				Times:     1,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
			},
		}, {
			name:      "auth token failure",
			path:      "/register/auth-token-failure",
			user:      fmt.Sprintf(testUserQuery["register"], "first name", "last name", "email@address.com", "username999", "password999"),
			expectErr: true,
			authHashData: &http_common.MockAuthData{
				OutputParam1: "hashed password",
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 1,
			},
			authGenJWTData: &http_common.MockAuthData{
				OutputErr: errors.New("auth token failure"),
				Times:     1,
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
			mockCassandra := mocks.NewMockCassandra(mockCtrl)
			mockRedis := mocks.NewMockRedis(mockCtrl)     // Not called.
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

			mockAuth.EXPECT().HashPassword(gomock.Any()).Return(
				testCase.authHashData.OutputParam1,
				testCase.authHashData.OutputErr,
			).Times(testCase.authHashData.Times)

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraCreateData.OutputParam,
				testCase.cassandraCreateData.OutputErr,
			).Times(testCase.cassandraCreateData.Times)

			mockAuth.EXPECT().GenerateJWT(gomock.Any()).Return(
				testCase.authGenJWTData.OutputParam1,
				testCase.authGenJWTData.OutputErr,
			).Times(testCase.authGenJWTData.Times)

			// Endpoint setup for test.
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrading, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(testCase.user))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, http.StatusOK, w.Code, "expected status codes do not match")

			response := map[string]any{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response), "failed to unmarshal response body")

			// Error is expected check to ensure one is set.
			if testCase.expectErr {
				verifyErrorReturned(t, response)
			} else {
				// Auth token is expected.
				verifyJWTReturned(t, response, "registerUser", testCase.authGenJWTData.OutputParam1.(*model_http.JWTAuthResponse))
			}
		})
	}
}

func TestMutationResolver_LoginUser(t *testing.T) {
	router := http_common.GetTestRouter()

	testCases := []struct {
		name              string
		path              string
		user              string
		expectErr         bool
		authCheckPwdData  *http_common.MockAuthData
		authGenJWTData    *http_common.MockAuthData
		cassandraReadData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty user",
			path:      "/login/empty-user",
			user:      fmt.Sprintf(testUserQuery["login"], "", ""),
			expectErr: true,
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
			authCheckPwdData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
			},
		}, {
			name:      "valid user",
			path:      "/login/valid-user",
			user:      fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr: false,
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       1,
			},
			authCheckPwdData: &http_common.MockAuthData{
				Times: 1,
			},
			authGenJWTData: &http_common.MockAuthData{
				OutputParam1: &model_http.JWTAuthResponse{
					Token:     "Test Token",
					Expires:   99,
					Threshold: 100,
				},
				Times: 1,
			},
		}, {
			name:      "database failure",
			path:      "/login/database-failure",
			user:      fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr: true,
			cassandraReadData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{Status: http.StatusNotFound},
				Times:     1,
			},
			authCheckPwdData: &http_common.MockAuthData{
				OutputParam1: "hashed password",
				Times:        0,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
			},
		}, {
			name:      "password check failure",
			path:      "/login/pwd-check-failure",
			user:      fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr: true,
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       1,
			},
			authCheckPwdData: &http_common.MockAuthData{
				OutputErr: errors.New("password hash failure"),
				Times:     1,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
			},
		}, {
			name:      "auth token failure",
			path:      "/login/auth-token-failure",
			user:      fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr: true,
			authCheckPwdData: &http_common.MockAuthData{
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       1,
			},
			authGenJWTData: &http_common.MockAuthData{
				OutputErr: errors.New("auth token failure"),
				Times:     1,
			},
		}, {
			name:      "deleted user",
			path:      "/login/deleted-user",
			user:      fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr: true,
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.User{
					UserAccount: &model_cassandra.UserAccount{
						UserLoginCredentials: model_cassandra.UserLoginCredentials{Password: "empty password"},
					},
					IsDeleted: true,
				},
				Times: 1,
			},
			authCheckPwdData: &http_common.MockAuthData{
				Times: 1,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
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
			mockCassandra := mocks.NewMockCassandra(mockCtrl)
			mockRedis := mocks.NewMockRedis(mockCtrl)     // Not called.
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraReadData.OutputParam,
				testCase.cassandraReadData.OutputErr,
			).Times(testCase.cassandraReadData.Times)

			mockAuth.EXPECT().CheckPassword(gomock.Any(), gomock.Any()).Return(
				testCase.authCheckPwdData.OutputErr,
			).Times(testCase.authCheckPwdData.Times)

			mockAuth.EXPECT().GenerateJWT(gomock.Any()).Return(
				testCase.authGenJWTData.OutputParam1,
				testCase.authGenJWTData.OutputErr,
			).Times(testCase.authGenJWTData.Times)

			// Endpoint setup for test.
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrading, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(testCase.user))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, http.StatusOK, w.Code, "expected status codes do not match")

			response := map[string]any{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response), "failed to unmarshal response body")

			// Error is expected check to ensure one is set.
			if testCase.expectErr {
				verifyErrorReturned(t, response)
			} else {
				// Auth token is expected.
				verifyJWTReturned(t, response, "loginUser", testCase.authGenJWTData.OutputParam1.(*model_http.JWTAuthResponse))
			}
		})
	}
}

func TestMutationResolver_RefreshToken(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := http_common.GetTestRouter()
	router.Use(GinContextToContextMiddleware())

	testCases := []struct {
		name                 string
		path                 string
		expectErr            bool
		authValidateJWTData  *http_common.MockAuthData
		authGenJWTData       *http_common.MockAuthData
		authRefThresholdData *http_common.MockAuthData
		cassandraReadData    *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/refresh/empty-token",
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
			authRefThresholdData: &http_common.MockAuthData{
				OutputParam1: int64(60),
				Times:        0,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
			},
		}, {
			name:      "valid token",
			path:      "/refresh/valid-token",
			expectErr: false,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				OutputParam2: time.Now().Add(-time.Duration(30) * time.Second).Unix(),
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       1,
			},
			authRefThresholdData: &http_common.MockAuthData{
				OutputParam1: int64(60),
				Times:        1,
			},
			authGenJWTData: &http_common.MockAuthData{
				OutputParam1: &model_http.JWTAuthResponse{
					Token:     "some valid token should be here",
					Expires:   9999999,
					Threshold: 5555,
				},
				Times: 1,
			},
		}, {
			name:      "valid token not expiring",
			path:      "/refresh/valid-token-not-expiring",
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				OutputParam2: time.Now().Add(time.Duration(3) * time.Minute).Unix(),
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       1,
			},
			authRefThresholdData: &http_common.MockAuthData{
				OutputParam1: int64(60),
				Times:        1,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
			},
		}, {
			name:      "invalid token",
			path:      "/refresh/invalid-token",
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				OutputParam2: time.Now().Add(-time.Duration(30) * time.Second).Unix(),
				OutputErr:    errors.New("validate JWT failure"),
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       0,
			},
			authRefThresholdData: &http_common.MockAuthData{
				OutputParam1: int64(60),
				Times:        0,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
			},
		}, {
			name:      "db failure",
			path:      "/refresh/db-failure",
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				OutputParam2: time.Now().Add(-time.Duration(30) * time.Second).Unix(),
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputErr: errors.New("db failure"),
				Times:     1,
			},
			authRefThresholdData: &http_common.MockAuthData{
				OutputParam1: int64(60),
				Times:        0,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
			},
		}, {
			name:      "deleted user",
			path:      "/refresh/deleted-user",
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.User{
					IsDeleted: true,
				},
				Times: 1,
			},
			authRefThresholdData: &http_common.MockAuthData{
				OutputParam1: int64(60),
				Times:        0,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 0,
			},
		}, {
			name:      "token generation failure",
			path:      "/refresh/token-generation-failure",
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				OutputParam2: time.Now().Add(-time.Duration(30) * time.Second).Unix(),
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       1,
			},
			authRefThresholdData: &http_common.MockAuthData{
				OutputParam1: int64(60),
				Times:        1,
			},
			authGenJWTData: &http_common.MockAuthData{
				OutputErr: errors.New("failed to generate token"),
				Times:     1,
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
			mockCassandra := mocks.NewMockCassandra(mockCtrl)
			mockRedis := mocks.NewMockRedis(mockCtrl)     // Not called.
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

			gomock.InOrder(
				// JWT check.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.OutputParam1,
					testCase.authValidateJWTData.OutputParam2,
					testCase.authValidateJWTData.OutputErr,
				).Times(testCase.authValidateJWTData.Times),

				// Database call for user record.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraReadData.OutputParam,
					testCase.cassandraReadData.OutputErr,
				).Times(testCase.cassandraReadData.Times),

				// Refresh threshold request.
				mockAuth.EXPECT().RefreshThreshold().Return(
					testCase.authRefThresholdData.OutputParam1,
				).Times(testCase.authRefThresholdData.Times),

				// Generate fresh JWT.
				mockAuth.EXPECT().GenerateJWT(gomock.Any()).Return(
					testCase.authGenJWTData.OutputParam1,
					testCase.authGenJWTData.OutputErr,
				).Times(testCase.authGenJWTData.Times),
			)

			// Endpoint setup for test.
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrading, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(testUserQuery["refresh"]))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "some valid auth token goes here")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, http.StatusOK, w.Code, "expected status codes do not match")

			response := map[string]any{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response), "failed to unmarshal response body")

			// Error is expected check to ensure one is set.
			if testCase.expectErr {
				verifyErrorReturned(t, response)
			} else {
				// Auth token is expected.
				verifyJWTReturned(t, response, "refreshToken", testCase.authGenJWTData.OutputParam1.(*model_http.JWTAuthResponse))
			}
		})
	}
}

func TestMutationResolver_DeleteUser(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := http_common.GetTestRouter()
	router.Use(GinContextToContextMiddleware())

	testCases := []struct {
		name                string
		path                string
		query               string
		expectErr           bool
		authValidateJWTData *http_common.MockAuthData
		authCheckPwdData    *http_common.MockAuthData
		cassandraReadData   *http_common.MockCassandraData
		cassandraDeleteData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty request",
			path:      "/delete/empty-request",
			query:     fmt.Sprintf(testUserQuery["delete"], "", "", ""),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
			authCheckPwdData: &http_common.MockAuthData{
				Times: 0,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "valid token",
			path:      "/delete/valid-request",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: false,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       1,
			},
			authCheckPwdData: &http_common.MockAuthData{
				Times: 1,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 1,
			},
		}, {
			name:      "invalid token",
			path:      "/delete/invalid-token-request",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				OutputErr:    errors.New("JWT failed authorization check"),
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       0,
			},
			authCheckPwdData: &http_common.MockAuthData{
				Times: 0,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "token and request username mismatch",
			path:      "/delete/token-and-request-username-mismatch",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username mismatch",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
			authCheckPwdData: &http_common.MockAuthData{
				Times: 0,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "db read failure",
			path:      "/delete/db-read-failure",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				OutputErr:   errors.New("db read failure"),
				Times:       1,
			},
			authCheckPwdData: &http_common.MockAuthData{
				Times: 0,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "already deleted",
			path:      "/delete/already-deleted",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.User{
					IsDeleted: true,
				},
				Times: 1,
			},
			authCheckPwdData: &http_common.MockAuthData{
				Times: 0,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "db delete failure",
			path:      "/delete/db-delete-failure",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       1,
			},
			authCheckPwdData: &http_common.MockAuthData{
				Times: 1,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				OutputErr: errors.New("db delete failure"),
				Times:     1,
			},
		}, {
			name:      "bad deletion confirmation",
			path:      "/delete/bad-deletion-confirmation",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "incorrect and incomplete confirmation"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
			authCheckPwdData: &http_common.MockAuthData{
				Times: 0,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "invalid password",
			path:      "/delete/valid-password",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "incorrect password", "username1"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       1,
			},
			authCheckPwdData: &http_common.MockAuthData{
				OutputErr: errors.New("password check failed"),
				Times:     1,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 0,
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
			mockCassandra := mocks.NewMockCassandra(mockCtrl)
			mockRedis := mocks.NewMockRedis(mockCtrl)     // Not called.
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

			authToken := xid.New().String()

			gomock.InOrder(
				// Authorization check.
				mockAuth.EXPECT().ValidateJWT(authToken).Return(
					testCase.authValidateJWTData.OutputParam1,
					testCase.authValidateJWTData.OutputParam2,
					testCase.authValidateJWTData.OutputErr,
				).Times(testCase.authValidateJWTData.Times),
				// DB read call.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraReadData.OutputParam,
					testCase.cassandraReadData.OutputErr,
				).Times(testCase.cassandraReadData.Times),
				// Password check.
				mockAuth.EXPECT().CheckPassword(gomock.Any(), gomock.Any()).Return(
					testCase.authCheckPwdData.OutputErr,
				).Times(testCase.authCheckPwdData.Times),
				// DB delete call.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraDeleteData.OutputParam,
					testCase.cassandraDeleteData.OutputErr,
				).Times(testCase.cassandraDeleteData.Times),
			)

			// Endpoint setup for test.
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrading, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(testCase.query))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", authToken)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, http.StatusOK, w.Code, "expected status codes do not match")

			response := map[string]any{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response), "failed to unmarshal response body")

			// Error is expected check to ensure one is set.
			if testCase.expectErr {
				verifyErrorReturned(t, response)
			} else {
				// Auth token is expected.
				data, ok := response["data"]
				require.True(t, ok, "data key expected but not set")
				confirmation := data.(map[string]any)["deleteUser"].(string)
				require.Equal(t, "account successfully deleted", confirmation, "confirmation message does not match expected")
			}
		})
	}
}

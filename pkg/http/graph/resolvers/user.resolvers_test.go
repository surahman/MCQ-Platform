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
	"github.com/surahman/mcq-platform/pkg/mocks"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
)

func TestMutationResolver_RegisterUser(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		user                string
		expectErr           bool
		authHashData        *mockAuthData
		authGenJWTData      *mockAuthData
		cassandraCreateData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty user",
			path:      "/register/empty-user",
			user:      fmt.Sprintf(testUserQuery["register"], "", "", "", "", ""),
			expectErr: true,
			authHashData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
			},
		}, {
			name: "valid user",
			path: "/register/valid-user",
			user: fmt.Sprintf(testUserQuery["register"], "first name", "last name", "email@address.com", "username999", "password999"),
			authHashData: &mockAuthData{
				outputParam1: "hashed password",
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 1,
			},
			authGenJWTData: &mockAuthData{
				outputParam1: &model_http.JWTAuthResponse{
					Token:     "Test Token",
					Expires:   99,
					Threshold: 100,
				},
				times: 1,
			},
		}, {
			name:      "password hash failure",
			path:      "/register/pwd-hash-failure",
			user:      fmt.Sprintf(testUserQuery["register"], "first name", "last name", "email@address.com", "username999", "password999"),
			expectErr: true,
			authHashData: &mockAuthData{
				outputParam1: "hashed password",
				outputErr:    errors.New("password hash failure"),
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
			},
		}, {
			name:      "database failure",
			path:      "/register/database-failure",
			user:      fmt.Sprintf(testUserQuery["register"], "first name", "last name", "email@address.com", "username999", "password999"),
			expectErr: true,
			authHashData: &mockAuthData{
				outputParam1: "hashed password",
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputErr: &cassandra.Error{Status: http.StatusNotFound},
				times:     1,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
			},
		}, {
			name:      "auth token failure",
			path:      "/register/auth-token-failure",
			user:      fmt.Sprintf(testUserQuery["register"], "first name", "last name", "email@address.com", "username999", "password999"),
			expectErr: true,
			authHashData: &mockAuthData{
				outputParam1: "hashed password",
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 1,
			},
			authGenJWTData: &mockAuthData{
				outputErr: errors.New("auth token failure"),
				times:     1,
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
				testCase.authHashData.outputParam1,
				testCase.authHashData.outputErr,
			).Times(testCase.authHashData.times)

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraCreateData.outputParam,
				testCase.cassandraCreateData.outputErr,
			).Times(testCase.cassandraCreateData.times)

			mockAuth.EXPECT().GenerateJWT(gomock.Any()).Return(
				testCase.authGenJWTData.outputParam1,
				testCase.authGenJWTData.outputErr,
			).Times(testCase.authGenJWTData.times)

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
				verifyJWTReturned(t, response, "registerUser", testCase.authGenJWTData.outputParam1.(*model_http.JWTAuthResponse))
			}
		})
	}
}

func TestMutationResolver_LoginUser(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name              string
		path              string
		user              string
		expectErr         bool
		authCheckPwdData  *mockAuthData
		authGenJWTData    *mockAuthData
		cassandraReadData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty user",
			path:      "/login/empty-user",
			user:      fmt.Sprintf(testUserQuery["login"], "", ""),
			expectErr: true,
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			authCheckPwdData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
			},
		}, {
			name:      "valid user",
			path:      "/login/valid-user",
			user:      fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr: false,
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       1,
			},
			authCheckPwdData: &mockAuthData{
				times: 1,
			},
			authGenJWTData: &mockAuthData{
				outputParam1: &model_http.JWTAuthResponse{
					Token:     "Test Token",
					Expires:   99,
					Threshold: 100,
				},
				times: 1,
			},
		}, {
			name:      "database failure",
			path:      "/login/database-failure",
			user:      fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr: true,
			cassandraReadData: &mockCassandraData{
				outputErr: &cassandra.Error{Status: http.StatusNotFound},
				times:     1,
			},
			authCheckPwdData: &mockAuthData{
				outputParam1: "hashed password",
				times:        0,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
			},
		}, {
			name:      "password check failure",
			path:      "/login/pwd-check-failure",
			user:      fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr: true,
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       1,
			},
			authCheckPwdData: &mockAuthData{
				outputErr: errors.New("password hash failure"),
				times:     1,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
			},
		}, {
			name:      "auth token failure",
			path:      "/login/auth-token-failure",
			user:      fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr: true,
			authCheckPwdData: &mockAuthData{
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       1,
			},
			authGenJWTData: &mockAuthData{
				outputErr: errors.New("auth token failure"),
				times:     1,
			},
		}, {
			name:      "deleted user",
			path:      "/login/deleted-user",
			user:      fmt.Sprintf(testUserQuery["login"], "username999", "password999"),
			expectErr: true,
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.User{
					UserAccount: &model_cassandra.UserAccount{
						UserLoginCredentials: model_cassandra.UserLoginCredentials{Password: "empty password"},
					},
					IsDeleted: true,
				},
				times: 1,
			},
			authCheckPwdData: &mockAuthData{
				times: 1,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
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
				testCase.cassandraReadData.outputParam,
				testCase.cassandraReadData.outputErr,
			).Times(testCase.cassandraReadData.times)

			mockAuth.EXPECT().CheckPassword(gomock.Any(), gomock.Any()).Return(
				testCase.authCheckPwdData.outputErr,
			).Times(testCase.authCheckPwdData.times)

			mockAuth.EXPECT().GenerateJWT(gomock.Any()).Return(
				testCase.authGenJWTData.outputParam1,
				testCase.authGenJWTData.outputErr,
			).Times(testCase.authGenJWTData.times)

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
				verifyJWTReturned(t, response, "loginUser", testCase.authGenJWTData.outputParam1.(*model_http.JWTAuthResponse))
			}
		})
	}
}

func TestMutationResolver_RefreshToken(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := getRouter()
	router.Use(GinContextToContextMiddleware())

	testCases := []struct {
		name                 string
		path                 string
		expectErr            bool
		authValidateJWTData  *mockAuthData
		authGenJWTData       *mockAuthData
		authRefThresholdData *mockAuthData
		cassandraReadData    *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/refresh/empty-token",
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			authRefThresholdData: &mockAuthData{
				outputParam1: int64(60),
				times:        0,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
			},
		}, {
			name:      "valid token",
			path:      "/refresh/valid-token",
			expectErr: false,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				outputParam2: time.Now().Add(-time.Duration(30) * time.Second).Unix(),
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       1,
			},
			authRefThresholdData: &mockAuthData{
				outputParam1: int64(60),
				times:        1,
			},
			authGenJWTData: &mockAuthData{
				outputParam1: &model_http.JWTAuthResponse{
					Token:     "some valid token should be here",
					Expires:   9999999,
					Threshold: 5555,
				},
				times: 1,
			},
		}, {
			name:      "valid token not expiring",
			path:      "/refresh/valid-token-not-expiring",
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				outputParam2: time.Now().Add(-time.Duration(3) * time.Minute).Unix(),
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       1,
			},
			authRefThresholdData: &mockAuthData{
				outputParam1: int64(60),
				times:        1,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
			},
		}, {
			name:      "invalid token",
			path:      "/refresh/invalid-token",
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				outputParam2: time.Now().Add(-time.Duration(30) * time.Second).Unix(),
				outputErr:    errors.New("validate JWT failure"),
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       0,
			},
			authRefThresholdData: &mockAuthData{
				outputParam1: int64(60),
				times:        0,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
			},
		}, {
			name:      "db failure",
			path:      "/refresh/db-failure",
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				outputParam2: time.Now().Add(-time.Duration(30) * time.Second).Unix(),
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputErr: errors.New("db failure"),
				times:     1,
			},
			authRefThresholdData: &mockAuthData{
				outputParam1: int64(60),
				times:        0,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
			},
		}, {
			name:      "deleted user",
			path:      "/refresh/deleted-user",
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.User{
					IsDeleted: true,
				},
				times: 1,
			},
			authRefThresholdData: &mockAuthData{
				outputParam1: int64(60),
				times:        0,
			},
			authGenJWTData: &mockAuthData{
				times: 0,
			},
		}, {
			name:      "token generation failure",
			path:      "/refresh/token-generation-failure",
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				outputParam2: time.Now().Add(-time.Duration(30) * time.Second).Unix(),
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       1,
			},
			authRefThresholdData: &mockAuthData{
				outputParam1: int64(60),
				times:        1,
			},
			authGenJWTData: &mockAuthData{
				outputErr: errors.New("failed to generate token"),
				times:     1,
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
					testCase.authValidateJWTData.outputParam1,
					testCase.authValidateJWTData.outputParam2,
					testCase.authValidateJWTData.outputErr,
				).Times(testCase.authValidateJWTData.times),

				// Database call for user record.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraReadData.outputParam,
					testCase.cassandraReadData.outputErr,
				).Times(testCase.cassandraReadData.times),

				// Refresh threshold request.
				mockAuth.EXPECT().RefreshThreshold().Return(
					testCase.authRefThresholdData.outputParam1,
				).Times(testCase.authRefThresholdData.times),

				// Generate fresh JWT.
				mockAuth.EXPECT().GenerateJWT(gomock.Any()).Return(
					testCase.authGenJWTData.outputParam1,
					testCase.authGenJWTData.outputErr,
				).Times(testCase.authGenJWTData.times),
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
				verifyJWTReturned(t, response, "refreshToken", testCase.authGenJWTData.outputParam1.(*model_http.JWTAuthResponse))
			}
		})
	}
}

func TestMutationResolver_DeleteUser(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := getRouter()
	router.Use(GinContextToContextMiddleware())

	testCases := []struct {
		name                string
		path                string
		query               string
		expectErr           bool
		authValidateJWTData *mockAuthData
		authCheckPwdData    *mockAuthData
		cassandraReadData   *mockCassandraData
		cassandraDeleteData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty request",
			path:      "/delete/empty-request",
			query:     fmt.Sprintf(testUserQuery["delete"], "", "", ""),
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			authCheckPwdData: &mockAuthData{
				times: 0,
			},
			cassandraDeleteData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "valid token",
			path:      "/delete/valid-request",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: false,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       1,
			},
			authCheckPwdData: &mockAuthData{
				times: 1,
			},
			cassandraDeleteData: &mockCassandraData{
				times: 1,
			},
		}, {
			name:      "invalid token",
			path:      "/delete/invalid-token-request",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				outputErr:    errors.New("JWT failed authorization check"),
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       0,
			},
			authCheckPwdData: &mockAuthData{
				times: 0,
			},
			cassandraDeleteData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "token and request username mismatch",
			path:      "/delete/token-and-request-username-mismatch",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username mismatch",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			authCheckPwdData: &mockAuthData{
				times: 0,
			},
			cassandraDeleteData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "db read failure",
			path:      "/delete/db-read-failure",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				outputErr:   errors.New("db read failure"),
				times:       1,
			},
			authCheckPwdData: &mockAuthData{
				times: 0,
			},
			cassandraDeleteData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "already deleted",
			path:      "/delete/already-deleted",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.User{
					IsDeleted: true,
				},
				times: 1,
			},
			authCheckPwdData: &mockAuthData{
				times: 0,
			},
			cassandraDeleteData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "db delete failure",
			path:      "/delete/db-delete-failure",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "username1"),
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       1,
			},
			authCheckPwdData: &mockAuthData{
				times: 1,
			},
			cassandraDeleteData: &mockCassandraData{
				outputErr: errors.New("db delete failure"),
				times:     1,
			},
		}, {
			name:      "bad deletion confirmation",
			path:      "/delete/bad-deletion-confirmation",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "password", "incorrect and incomplete confirmation"),
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			authCheckPwdData: &mockAuthData{
				times: 0,
			},
			cassandraDeleteData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "invalid password",
			path:      "/delete/valid-password",
			query:     fmt.Sprintf(testUserQuery["delete"], "username1", "incorrect password", "username1"),
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "username1",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       1,
			},
			authCheckPwdData: &mockAuthData{
				outputErr: errors.New("password check failed"),
				times:     1,
			},
			cassandraDeleteData: &mockCassandraData{
				times: 0,
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
					testCase.authValidateJWTData.outputParam1,
					testCase.authValidateJWTData.outputParam2,
					testCase.authValidateJWTData.outputErr,
				).Times(testCase.authValidateJWTData.times),
				// DB read call.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraReadData.outputParam,
					testCase.cassandraReadData.outputErr,
				).Times(testCase.cassandraReadData.times),
				// Password check.
				mockAuth.EXPECT().CheckPassword(gomock.Any(), gomock.Any()).Return(
					testCase.authCheckPwdData.outputErr,
				).Times(testCase.authCheckPwdData.times),
				// DB delete call.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraDeleteData.outputParam,
					testCase.cassandraDeleteData.outputErr,
				).Times(testCase.cassandraDeleteData.times),
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

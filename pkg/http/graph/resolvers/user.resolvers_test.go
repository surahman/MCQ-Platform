package graphql_resolvers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/mocks"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
)

func TestMutationResolver_RegisterUser(t *testing.T) {
	router := getRouter()

	emptyUser := `{
    "query": "mutation { registerUser(input: { firstname: \"\", lastname:\"\", email: \"\", userLoginCredentials: { username:\"\", password: \"\" } }) { token, expires, threshold }}"
}`

	validUser := `{
    "query": "mutation { registerUser(input: { firstname: \"first name\", lastname:\"last name\", email: \"email@address.com\", userLoginCredentials: { username:\"username999\", password: \"password999\" } }) { token, expires, threshold }}"
}`

	testCases := []struct {
		name                string
		path                string
		user                *string
		expectErr           bool
		authHashData        *mockAuthData
		authGenJWTData      *mockAuthData
		cassandraCreateData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty user",
			path:      "/register/empty-user",
			user:      &emptyUser,
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
			user: &validUser,
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
			user:      &validUser,
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
			user:      &validUser,
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
			user:      &validUser,
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

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(*testCase.user))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, http.StatusOK, w.Code, "expected status codes do not match")

			response := map[string]any{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response), "failed to unmarshal response body")

			// Error is expected check to ensure one is set.
			if testCase.expectErr {
				value, ok := response["data"]
				require.True(t, ok, "data key expected but not set")
				require.Nil(t, value, "data value should be set to nil")

				value, ok = response["errors"]
				require.True(t, ok, "error key expected but not set")
				require.NotNil(t, value, "error value should not be nil")
			} else {
				// Auth token is expected.
				data, ok := response["data"]
				require.True(t, ok, "data key expected but not set")

				authToken := model_http.JWTAuthResponse{}
				jsonStr, err := json.Marshal(data.(map[string]any)["registerUser"])
				require.NoError(t, err, "failed to generate JSON string")
				require.NoError(t, json.Unmarshal([]byte(jsonStr), &authToken), "failed to unmarshall to JWT Auth Response")
				require.True(t, reflect.DeepEqual(*testCase.authGenJWTData.outputParam1.(*model_http.JWTAuthResponse), authToken), "auth tokens did not match")
			}
		})
	}
}

func TestMutationResolver_LoginUser(t *testing.T) {
	router := getRouter()

	emptyUser := `{
    "query": "mutation { loginUser(input: { username:\"\", password: \"\" }) { token, expires, threshold }}"
}`

	validUser := `{
    "query": "mutation { loginUser(input: { username:\"username999\", password: \"password999\" }) { token, expires, threshold }}"
}`

	testCases := []struct {
		name              string
		path              string
		user              *string
		expectErr         bool
		authCheckPwdData  *mockAuthData
		authGenJWTData    *mockAuthData
		cassandraReadData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty user",
			path:      "/login/empty-user",
			user:      &emptyUser,
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
			user:      &validUser,
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
			user:      &validUser,
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
			user:      &validUser,
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
			user:      &validUser,
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
			user:      &validUser,
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

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(*testCase.user))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, http.StatusOK, w.Code, "expected status codes do not match")

			response := map[string]any{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response), "failed to unmarshal response body")

			// Error is expected check to ensure one is set.
			if testCase.expectErr {
				value, ok := response["data"]
				require.True(t, ok, "data key expected but not set")
				require.Nil(t, value, "data value should be set to nil")

				value, ok = response["errors"]
				require.True(t, ok, "error key expected but not set")
				require.NotNil(t, value, "error value should not be nil")
			} else {
				// Auth token is expected.
				data, ok := response["data"]
				require.True(t, ok, "data key expected but not set")

				authToken := model_http.JWTAuthResponse{}
				jsonStr, err := json.Marshal(data.(map[string]any)["loginUser"])
				require.NoError(t, err, "failed to generate JSON string")
				require.NoError(t, json.Unmarshal([]byte(jsonStr), &authToken), "failed to unmarshall to JWT Auth Response")
				require.True(t, reflect.DeepEqual(*testCase.authGenJWTData.outputParam1.(*model_http.JWTAuthResponse), authToken), "auth tokens did not match")
			}
		})
	}
}

func TestMutationResolver_RefreshToken(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := getRouter()
	router.Use(GinContextToContextMiddleware())

	refreshTokenQuery := `{ "query": "mutation { refreshToken() { token expires threshold }}" }`

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
			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(refreshTokenQuery))
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
				value, ok := response["data"]
				require.True(t, ok, "data key expected but not set")
				require.Nil(t, value, "data value should be set to nil")

				value, ok = response["errors"]
				require.True(t, ok, "error key expected but not set")
				require.NotNil(t, value, "error value should not be nil")
			} else {
				// Auth token is expected.
				data, ok := response["data"]
				require.True(t, ok, "data key expected but not set")

				authToken := model_http.JWTAuthResponse{}
				jsonStr, err := json.Marshal(data.(map[string]any)["refreshToken"])
				require.NoError(t, err, "failed to generate JSON string")
				require.NoError(t, json.Unmarshal([]byte(jsonStr), &authToken), "failed to unmarshall to JWT Auth Response")
				require.True(t, reflect.DeepEqual(*testCase.authGenJWTData.outputParam1.(*model_http.JWTAuthResponse), authToken), "auth tokens did not match")
			}
		})
	}
}

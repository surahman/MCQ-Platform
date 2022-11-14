package http_handlers

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
	"github.com/surahman/mcq-platform/pkg/constants"
	http_common "github.com/surahman/mcq-platform/pkg/http"
	"github.com/surahman/mcq-platform/pkg/mocks"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
)

func TestRegisterUser(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		expectedStatus      int
		user                *model_cassandra.UserAccount
		authHashData        *http_common.MockAuthData
		authGenJWTData      *http_common.MockAuthData
		cassandraCreateData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty user",
			path:           "/register/empty-user",
			expectedStatus: http.StatusBadRequest,
			user:           &model_cassandra.UserAccount{},
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
			name:           "valid user",
			path:           "/register/valid-user",
			expectedStatus: http.StatusOK,
			user:           testUserData["username1"].UserAccount,
			authHashData: &http_common.MockAuthData{
				OutputParam1: "hashed password",
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 1,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 1,
			},
		}, {
			name:           "password hash failure",
			path:           "/register/pwd-hash-failure",
			expectedStatus: http.StatusInternalServerError,
			user:           testUserData["username1"].UserAccount,
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
			name:           "database failure",
			path:           "/register/database-failure",
			expectedStatus: http.StatusNotFound,
			user:           testUserData["username1"].UserAccount,
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
			name:           "auth token failure",
			path:           "/register/auth-token-failure",
			expectedStatus: http.StatusInternalServerError,
			user:           testUserData["username1"].UserAccount,
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

			user := testCase.user
			userJson, err := json.Marshal(&user)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

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
			router.POST(testCase.path, RegisterUser(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBuffer(userJson))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}

func TestLoginUser(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name              string
		path              string
		expectedStatus    int
		user              *model_cassandra.UserLoginCredentials
		authCheckPwdData  *http_common.MockAuthData
		authGenJWTData    *http_common.MockAuthData
		cassandraReadData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty user",
			path:           "/login/empty-user",
			expectedStatus: http.StatusBadRequest,
			user:           &model_cassandra.UserLoginCredentials{},
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
			name:           "valid user",
			path:           "/login/valid-user",
			expectedStatus: http.StatusOK,
			user:           &testUserData["username1"].UserLoginCredentials,
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: testUserData["username1"],
				Times:       1,
			},
			authCheckPwdData: &http_common.MockAuthData{
				Times: 1,
			},
			authGenJWTData: &http_common.MockAuthData{
				Times: 1,
			},
		}, {
			name:           "database failure",
			path:           "/login/database-failure",
			expectedStatus: http.StatusForbidden,
			user:           &testUserData["username1"].UserLoginCredentials,
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
			name:           "password check failure",
			path:           "/login/pwd-check-failure",
			expectedStatus: http.StatusForbidden,
			user:           &testUserData["username1"].UserLoginCredentials,
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
			name:           "auth token failure",
			path:           "/login/auth-token-failure",
			expectedStatus: http.StatusInternalServerError,
			user:           &testUserData["username1"].UserLoginCredentials,
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
			name:           "deleted user",
			path:           "/login/deleted-user",
			expectedStatus: http.StatusForbidden,
			user:           &testUserData["username1"].UserLoginCredentials,
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

			user := testCase.user
			userJson, err := json.Marshal(&user)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

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
			router.POST(testCase.path, LoginUser(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBuffer(userJson))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}

func TestLoginRefresh(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                 string
		path                 string
		expectedStatus       int
		authValidateJWTData  *http_common.MockAuthData
		authGenJWTData       *http_common.MockAuthData
		authRefThresholdData *http_common.MockAuthData
		cassandraReadData    *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/refresh/empty-token",
			expectedStatus: http.StatusForbidden,
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
			name:           "valid token",
			path:           "/refresh/valid-token",
			expectedStatus: http.StatusOK,
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
				Times: 1,
			},
		}, {
			name:           "valid token not expiring",
			path:           "/refresh/valid-token-not-expiring",
			expectedStatus: http.StatusNotExtended,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "username1",
				OutputParam2: time.Now().Add(-time.Duration(3) * time.Minute).Unix(),
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
			name:           "invalid token",
			path:           "/refresh/invalid-token",
			expectedStatus: http.StatusForbidden,
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
			name:           "db failure",
			path:           "/refresh/db-failure",
			expectedStatus: http.StatusInternalServerError,
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
			name:           "deleted user",
			path:           "/refresh/deleted-user",
			expectedStatus: http.StatusForbidden,
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
			name:           "token generation failure",
			path:           "/refresh/token-generation-failure",
			expectedStatus: http.StatusInternalServerError,
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

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraReadData.OutputParam,
				testCase.cassandraReadData.OutputErr,
			).Times(testCase.cassandraReadData.Times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.OutputParam1,
				testCase.authValidateJWTData.OutputParam2,
				testCase.authValidateJWTData.OutputErr,
			).Times(testCase.authValidateJWTData.Times)

			mockAuth.EXPECT().GenerateJWT(gomock.Any()).Return(
				testCase.authGenJWTData.OutputParam1,
				testCase.authGenJWTData.OutputErr,
			).Times(testCase.authGenJWTData.Times)

			mockAuth.EXPECT().RefreshThreshold().Return(
				testCase.authRefThresholdData.OutputParam1,
			).Times(testCase.authRefThresholdData.Times)

			// Endpoint setup for test.
			router.POST(testCase.path, LoginRefresh(zapLogger, mockAuth, mockCassandra, "Authorization"))
			req, _ := http.NewRequest("POST", testCase.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}

func TestDeleteUser(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		expectedStatus      int
		deleteRequest       *model_http.DeleteUserRequest
		authValidateJWTData *http_common.MockAuthData
		authCheckPwdData    *http_common.MockAuthData
		cassandraReadData   *http_common.MockCassandraData
		cassandraDeleteData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty request",
			path:           "/delete/empty-request",
			expectedStatus: http.StatusBadRequest,
			deleteRequest:  &model_http.DeleteUserRequest{},
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
			name:           "valid token",
			path:           "/delete/valid-request",
			expectedStatus: http.StatusOK,
			deleteRequest: &model_http.DeleteUserRequest{
				UserLoginCredentials: model_cassandra.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.GetDeleteUserAccountConfirmation(), "username1"),
			},
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
			name:           "token and request username mismatch",
			path:           "/delete/token-and-request-username-mismatch",
			expectedStatus: http.StatusForbidden,
			deleteRequest: &model_http.DeleteUserRequest{
				UserLoginCredentials: model_cassandra.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.GetDeleteUserAccountConfirmation(), "username1"),
			},
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
			name:           "db read failure",
			path:           "/delete/db-read-failure",
			expectedStatus: http.StatusInternalServerError,
			deleteRequest: &model_http.DeleteUserRequest{
				UserLoginCredentials: model_cassandra.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.GetDeleteUserAccountConfirmation(), "username1"),
			},
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
			name:           "already deleted",
			path:           "/delete/already-deleted",
			expectedStatus: http.StatusForbidden,
			deleteRequest: &model_http.DeleteUserRequest{
				UserLoginCredentials: model_cassandra.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.GetDeleteUserAccountConfirmation(), "username1"),
			},
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
			name:           "db delete failure",
			path:           "/delete/db-delete-failure",
			expectedStatus: http.StatusInternalServerError,
			deleteRequest: &model_http.DeleteUserRequest{
				UserLoginCredentials: model_cassandra.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.GetDeleteUserAccountConfirmation(), "username1"),
			},
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
			name:           "bad deletion confirmation",
			path:           "/delete/bad-deletion-confirmation",
			expectedStatus: http.StatusBadRequest,
			deleteRequest: &model_http.DeleteUserRequest{
				UserLoginCredentials: model_cassandra.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.GetDeleteUserAccountConfirmation(), "incorrect and incomplete confirmation"),
			},
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
			name:           "invalid password",
			path:           "/delete/valid-password",
			expectedStatus: http.StatusForbidden,
			deleteRequest: &model_http.DeleteUserRequest{
				UserLoginCredentials: model_cassandra.UserLoginCredentials{
					Username: "username1",
					Password: "password",
				},
				Confirmation: fmt.Sprintf(constants.GetDeleteUserAccountConfirmation(), "username1"),
			},
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

			requestJson, err := json.Marshal(&testCase.deleteRequest)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			authToken := xid.New().String()

			gomock.InOrder(
				// DB read call.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraReadData.OutputParam,
					testCase.cassandraReadData.OutputErr,
				).Times(testCase.cassandraReadData.Times),
				// DB delete call.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraDeleteData.OutputParam,
					testCase.cassandraDeleteData.OutputErr,
				).Times(testCase.cassandraDeleteData.Times),
			)

			mockAuth.EXPECT().ValidateJWT(authToken).Return(
				testCase.authValidateJWTData.OutputParam1,
				testCase.authValidateJWTData.OutputParam2,
				testCase.authValidateJWTData.OutputErr,
			).Times(testCase.authValidateJWTData.Times)

			mockAuth.EXPECT().CheckPassword(gomock.Any(), gomock.Any()).Return(
				testCase.authCheckPwdData.OutputErr,
			).Times(testCase.authCheckPwdData.Times)

			// Endpoint setup for test.
			router.DELETE(testCase.path, DeleteUser(zapLogger, mockAuth, mockCassandra, "Authorization"))
			req, _ := http.NewRequest("DELETE", testCase.path, bytes.NewBuffer(requestJson))
			req.Header.Set("Authorization", authToken)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}

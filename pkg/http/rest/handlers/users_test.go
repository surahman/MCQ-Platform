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
		authHashData        *mockAuthData
		authGenJWTData      *mockAuthData
		cassandraCreateData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty user",
			path:           "/register/empty-user",
			expectedStatus: http.StatusBadRequest,
			user:           &model_cassandra.UserAccount{},
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
			name:           "valid user",
			path:           "/register/valid-user",
			expectedStatus: http.StatusOK,
			user:           testUserData["username1"].UserAccount,
			authHashData: &mockAuthData{
				outputParam1: "hashed password",
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 1,
			},
			authGenJWTData: &mockAuthData{
				times: 1,
			},
		}, {
			name:           "password hash failure",
			path:           "/register/pwd-hash-failure",
			expectedStatus: http.StatusInternalServerError,
			user:           testUserData["username1"].UserAccount,
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
			name:           "database failure",
			path:           "/register/database-failure",
			expectedStatus: http.StatusNotFound,
			user:           testUserData["username1"].UserAccount,
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
			name:           "auth token failure",
			path:           "/register/auth-token-failure",
			expectedStatus: http.StatusInternalServerError,
			user:           testUserData["username1"].UserAccount,
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

			user := testCase.user
			userJson, err := json.Marshal(&user)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

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
		authCheckPwdData  *mockAuthData
		authGenJWTData    *mockAuthData
		cassandraReadData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty user",
			path:           "/login/empty-user",
			expectedStatus: http.StatusBadRequest,
			user:           &model_cassandra.UserLoginCredentials{},
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
			name:           "valid user",
			path:           "/login/valid-user",
			expectedStatus: http.StatusOK,
			user:           &testUserData["username1"].UserLoginCredentials,
			cassandraReadData: &mockCassandraData{
				outputParam: testUserData["username1"],
				times:       1,
			},
			authCheckPwdData: &mockAuthData{
				times: 1,
			},
			authGenJWTData: &mockAuthData{
				times: 1,
			},
		}, {
			name:           "database failure",
			path:           "/login/database-failure",
			expectedStatus: http.StatusForbidden,
			user:           &testUserData["username1"].UserLoginCredentials,
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
			name:           "password check failure",
			path:           "/login/pwd-check-failure",
			expectedStatus: http.StatusForbidden,
			user:           &testUserData["username1"].UserLoginCredentials,
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
			name:           "auth token failure",
			path:           "/login/auth-token-failure",
			expectedStatus: http.StatusInternalServerError,
			user:           &testUserData["username1"].UserLoginCredentials,
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
			name:           "deleted user",
			path:           "/login/deleted-user",
			expectedStatus: http.StatusForbidden,
			user:           &testUserData["username1"].UserLoginCredentials,
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

			user := testCase.user
			userJson, err := json.Marshal(&user)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

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
		authValidateJWTData  *mockAuthData
		authGenJWTData       *mockAuthData
		authRefThresholdData *mockAuthData
		cassandraReadData    *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/refresh/empty-token",
			expectedStatus: http.StatusForbidden,
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
			name:           "valid token",
			path:           "/refresh/valid-token",
			expectedStatus: http.StatusOK,
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
				times: 1,
			},
		}, {
			name:           "valid token not expiring",
			path:           "/refresh/valid-token-not-expiring",
			expectedStatus: http.StatusNotExtended,
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
			name:           "invalid token",
			path:           "/refresh/invalid-token",
			expectedStatus: http.StatusForbidden,
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
			name:           "db failure",
			path:           "/refresh/db-failure",
			expectedStatus: http.StatusInternalServerError,
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
			name:           "deleted user",
			path:           "/refresh/deleted-user",
			expectedStatus: http.StatusForbidden,
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
			name:           "token generation failure",
			path:           "/refresh/token-generation-failure",
			expectedStatus: http.StatusInternalServerError,
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

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraReadData.outputParam,
				testCase.cassandraReadData.outputErr,
			).Times(testCase.cassandraReadData.times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

			mockAuth.EXPECT().GenerateJWT(gomock.Any()).Return(
				testCase.authGenJWTData.outputParam1,
				testCase.authGenJWTData.outputErr,
			).Times(testCase.authGenJWTData.times)

			mockAuth.EXPECT().RefreshThreshold().Return(
				testCase.authRefThresholdData.outputParam1,
			).Times(testCase.authRefThresholdData.times)

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
		authValidateJWTData *mockAuthData
		authCheckPwdData    *mockAuthData
		cassandraReadData   *mockCassandraData
		cassandraDeleteData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty request",
			path:           "/delete/empty-request",
			expectedStatus: http.StatusBadRequest,
			deleteRequest:  &model_http.DeleteUserRequest{},
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

			requestJson, err := json.Marshal(&testCase.deleteRequest)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			authToken := xid.New().String()

			gomock.InOrder(
				// DB read call.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraReadData.outputParam,
					testCase.cassandraReadData.outputErr,
				).Times(testCase.cassandraReadData.times),
				// DB delete call.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraDeleteData.outputParam,
					testCase.cassandraDeleteData.outputErr,
				).Times(testCase.cassandraDeleteData.times),
			)

			mockAuth.EXPECT().ValidateJWT(authToken).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

			mockAuth.EXPECT().CheckPassword(gomock.Any(), gomock.Any()).Return(
				testCase.authCheckPwdData.outputErr,
			).Times(testCase.authCheckPwdData.times)

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

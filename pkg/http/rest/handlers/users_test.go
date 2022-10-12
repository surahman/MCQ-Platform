package http_handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/mocks"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
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
				outputParam: "",
				outputErr:   nil,
				times:       0,
			},
			cassandraCreateData: &mockCassandraData{times: 0},
			authGenJWTData: &mockAuthData{
				outputParam: nil,
				outputErr:   nil,
				times:       0,
			},
		}, {
			name:           "valid user",
			path:           "/register/valid-user",
			expectedStatus: http.StatusOK,
			user:           testUserData["username1"].UserAccount,
			authHashData: &mockAuthData{
				outputParam: "hashed password",
				outputErr:   nil,
				times:       1,
			},
			cassandraCreateData: &mockCassandraData{
				outputParam: nil,
				outputErr:   nil,
				times:       1,
			},
			authGenJWTData: &mockAuthData{
				outputParam: nil,
				outputErr:   nil,
				times:       1,
			},
		}, {
			name:           "password hash failure",
			path:           "/register/pwd-hash-failure",
			expectedStatus: http.StatusInternalServerError,
			user:           testUserData["username1"].UserAccount,
			authHashData: &mockAuthData{
				outputParam: "hashed password",
				outputErr:   errors.New("password hash failure"),
				times:       1,
			},
			cassandraCreateData: &mockCassandraData{
				outputParam: nil,
				outputErr:   nil,
				times:       0,
			},
			authGenJWTData: &mockAuthData{
				outputParam: nil,
				outputErr:   nil,
				times:       0,
			},
		}, {
			name:           "database failure",
			path:           "/register/database-failure",
			expectedStatus: http.StatusNotFound,
			user:           testUserData["username1"].UserAccount,
			authHashData: &mockAuthData{
				outputParam: "hashed password",
				outputErr:   nil,
				times:       1,
			},
			cassandraCreateData: &mockCassandraData{
				outputParam: nil,
				outputErr:   &cassandra.Error{Status: http.StatusNotFound},
				times:       1,
			},
			authGenJWTData: &mockAuthData{
				outputParam: nil,
				outputErr:   nil,
				times:       0,
			},
		}, {
			name:           "auth token failure",
			path:           "/register/auth-token-failure",
			expectedStatus: http.StatusInternalServerError,
			user:           testUserData["username1"].UserAccount,
			authHashData: &mockAuthData{
				outputParam: "hashed password",
				outputErr:   nil,
				times:       1,
			},
			cassandraCreateData: &mockCassandraData{
				outputParam: nil,
				outputErr:   nil,
				times:       1,
			},
			authGenJWTData: &mockAuthData{
				outputParam: nil,
				outputErr:   errors.New("auth token failure"),
				times:       1,
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
				testCase.authHashData.outputParam,
				testCase.authHashData.outputErr,
			).Times(testCase.authHashData.times)

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraCreateData.outputParam,
				testCase.cassandraCreateData.outputErr,
			).Times(testCase.cassandraCreateData.times)

			mockAuth.EXPECT().GenerateJWT(gomock.Any()).Return(
				testCase.authGenJWTData.outputParam,
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

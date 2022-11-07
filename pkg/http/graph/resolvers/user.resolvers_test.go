package graphql_resolvers

import (
	"bytes"
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
	// ABORT TESTING
	t.SkipNow()

	router := getRouter()

	query := `
mutation {
  registerUser(input: {
    firstname: "first name", 
  	lastname:"last name",
    email: "email@address.com",
    userLoginCredentials: {
      username:"username999",
    	password: "password999"
    }
  }) {
    token,
    expires,
    threshold
  }
}`

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
			mockRedis := mocks.NewMockRedis(mockCtrl)     // Not called.
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

			// user := testCase.user
			// userJson, err := json.Marshal(&user)
			// require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

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
			router.POST(testCase.path, QueryHandler(mockAuth, mockRedis, mockCassandra, mockGrading, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(query))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")
		})
	}
}

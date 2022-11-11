package graphql_resolvers

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
)

func TestMutationResolver_CreateQuiz(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := getRouter()
	router.Use(GinContextToContextMiddleware())

	testCases := []struct {
		name                string
		path                string
		query               string
		expectErr           bool
		authValidateJWTData *mockAuthData
		cassandraCreateData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/create/empty-token",
			query:     testQuizQuery["create_valid"],
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "empty quiz",
			path:      "/create/empty-quiz",
			query:     testQuizQuery["create_empty"],
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "valid quiz",
			path:      "/create/valid-quiz",
			query:     testQuizQuery["create_valid"],
			expectErr: false,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 1,
			},
		}, {
			name:      "invalid quiz",
			path:      "/create/invalid-quiz",
			query:     testQuizQuery["create_invalid"],
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "db failure internal",
			path:      "/create/db failure internal",
			query:     testQuizQuery["create_valid"],
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				times: 1,
			},
		}, {
			name:      "db failure conflict",
			path:      "/create/db failure conflict",
			query:     testQuizQuery["create_valid"],
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusConflict,
				},
				times: 1,
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
				// Check JWT.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.outputParam1,
					testCase.authValidateJWTData.outputParam2,
					testCase.authValidateJWTData.outputErr,
				).Times(testCase.authValidateJWTData.times),

				// Store in database.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraCreateData.outputParam,
					testCase.cassandraCreateData.outputErr,
				).Times(testCase.cassandraCreateData.times),
			)

			// Endpoint setup for test.
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrading, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(testCase.query))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "some auth token")
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
				// Quiz ID is expected.
				data, ok := response["data"]
				require.True(t, ok, "data key expected but not set")
				quizID := data.(map[string]any)["createQuiz"].(string)
				require.True(t, len(quizID) > 0, "no quiz id returned")
			}
		})
	}
}

package graphql_resolvers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/mocks"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
)

func TestMetadataResolver_QuizID(t *testing.T) {
	resolver := metadataResolver{}
	testCases := []struct {
		name        string
		metadata    *model_http.Metadata
		expectErr   require.ErrorAssertionFunc
		expectedLen int
	}{
		// ----- test cases start ----- //
		{
			name:        "nil metadata",
			metadata:    nil,
			expectErr:   require.Error,
			expectedLen: 0,
		}, {
			name:        "some metadata",
			metadata:    &model_http.Metadata{},
			expectErr:   require.NoError,
			expectedLen: 36,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			quizID, err := resolver.QuizID(context.TODO(), testCase.metadata)
			testCase.expectErr(t, err, "error expectation failed")
			require.Equal(t, testCase.expectedLen, len(quizID), "UUID not returned")
		})
	}
}

func TestQueryResolver_GetScore(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := getRouter()
	router.Use(GinContextToContextMiddleware())

	quizUUID := gocql.TimeUUID()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectErr           bool
		authValidateJWTData *mockAuthData
		cassandraReadData   *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/score/empty-token/",
			quizId:    quizUUID.String(),
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "invalid quiz id",
			path:      "/score/invalid-quiz-id",
			quizId:    "not a valid uuid",
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "db read not found",
			path:      "/score/db-read-not-found/",
			quizId:    quizUUID.String(),
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusNotFound,
				},
				times: 1,
			},
		}, {
			name:      "success",
			path:      "/score/success/",
			quizId:    quizUUID.String(),
			expectErr: false,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.Response{
					Username:     "mock response card",
					Score:        99.99,
					QuizResponse: nil,
					QuizID:       quizUUID,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)    // Not called.
			mockGrader := mocks.NewMockGrading(mockCtrl) // Not called.

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraReadData.outputParam,
				testCase.cassandraReadData.outputErr,
			).Times(testCase.cassandraReadData.times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

			// Endpoint setup for test.
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrader, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(fmt.Sprintf(testScoresQuery["score"], testCase.quizId)))
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
				expectedResponse := testCase.cassandraReadData.outputParam.(*model_cassandra.Response)

				data, ok := response["data"]
				require.True(t, ok, "data key expected but not set")

				responseMap := data.(map[string]any)["getScore"]

				actualScore := responseMap.(map[string]any)["Score"].(float64)
				require.InDelta(t, expectedResponse.Score, actualScore, 0.01, "returned score mismatch")

				actualUUID := responseMap.(map[string]any)["QuizID"]
				require.Equal(t, expectedResponse.QuizID.String(), actualUUID, "quiz id mismatch")
			}
		})
	}
}

func TestQueryResolver_GetStats(t *testing.T) {

}

func TestQueryResolver_prepareStatsResponse(t *testing.T) {
	
}

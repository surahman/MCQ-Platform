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
	http_common "github.com/surahman/mcq-platform/pkg/http"
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
		authValidateJWTData *http_common.MockAuthData
		cassandraReadData   *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/score/empty-token/",
			quizId:    quizUUID.String(),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "invalid quiz id",
			path:      "/score/invalid-quiz-id",
			quizId:    "not a valid uuid",
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "db read not found",
			path:      "/score/db-read-not-found/",
			quizId:    quizUUID.String(),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusNotFound,
				},
				Times: 1,
			},
		}, {
			name:      "success",
			path:      "/score/success/",
			quizId:    quizUUID.String(),
			expectErr: false,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Response{
					Username:     "mock response card",
					Score:        99.99,
					QuizResponse: nil,
					QuizID:       quizUUID,
				},
				Times: 1,
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
				testCase.cassandraReadData.OutputParam,
				testCase.cassandraReadData.OutputErr,
			).Times(testCase.cassandraReadData.Times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.OutputParam1,
				testCase.authValidateJWTData.OutputParam2,
				testCase.authValidateJWTData.OutputErr,
			).Times(testCase.authValidateJWTData.Times)

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
				expectedResponse := testCase.cassandraReadData.OutputParam.(*model_cassandra.Response)

				data, ok := response["data"]
				require.True(t, ok, "data key expected but not set")

				responseMap := data.(map[string]any)["getScore"]

				actualScore := responseMap.(map[string]any)["score"].(float64)
				require.InDelta(t, expectedResponse.Score, actualScore, 0.01, "returned score mismatch")

				actualUUID := responseMap.(map[string]any)["quizID"]
				require.Equal(t, expectedResponse.QuizID.String(), actualUUID, "quiz id mismatch")
			}
		})
	}
}

func TestQueryResolver_GetStats(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := getRouter()
	router.Use(GinContextToContextMiddleware())

	quizUUID := gocql.TimeUUID().String()

	testCases := []struct {
		name                string
		path                string
		query               string
		expectCursor        string
		expectPageSize      int
		expectNumRecords    int
		expectErr           bool
		authValidateJWTData *http_common.MockAuthData
		authDecryptData     *http_common.MockAuthData
		cassandraStatsData  *http_common.MockCassandraData
		authEncryptData     *http_common.MockAuthData
	}{
		// ----- test cases start ----- //
		{
			name:      "bad uuid",
			path:      "/stats-page/bad-uuid/",
			query:     fmt.Sprintf(testScoresQuery["stats"], "face palm", 3, "PaGeCuRs0R"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			authDecryptData: &http_common.MockAuthData{Times: 0},
			cassandraStatsData: &http_common.MockCassandraData{
				Times: 0,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:      "empty token",
			path:      "/stats-page/empty-token/",
			query:     fmt.Sprintf(testScoresQuery["stats"], quizUUID, 3, "PaGeCuRs0R"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 0},
			cassandraStatsData: &http_common.MockCassandraData{
				Times: 0,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:      "db read invalid user",
			path:      "/stats-page/failed-db-read-invalid-user/",
			query:     fmt.Sprintf(testScoresQuery["stats"], quizUUID, 3, "PaGeCuRs0R"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 1},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					Records: []*model_cassandra.Response{{Author: "UNexpected-username"}},
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:      "db read no records",
			path:      "/stats-page/failed-db-read-no-records/",
			query:     fmt.Sprintf(testScoresQuery["stats"], quizUUID, 3, "PaGeCuRs0R"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 1},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					Records: []*model_cassandra.Response{},
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:             "db quiz read valid user invalid page size",
			path:             "/stats-page/failed-db-quiz-read-valid-user-invalid-page-size/",
			query:            fmt.Sprintf(testScoresQuery["stats"], quizUUID, -1, "PaGeCuRs0R"),
			expectNumRecords: 1,
			expectErr:        false,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{
				OutputParam1: []byte("PaGeCuRs0R"),
				Times:        1,
			},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					Records: []*model_cassandra.Response{{Author: "expected-username"}}},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:      "db stat read failure",
			path:      "/stats-page/db-stat-read-failure/",
			query:     fmt.Sprintf(testScoresQuery["stats"], quizUUID, 3, "PaGeCuRs0R"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 1},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Status: http.StatusInternalServerError,
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:      "prepare response failure",
			path:      "/stats-page/prepare-response-failure/",
			query:     fmt.Sprintf(testScoresQuery["stats"], quizUUID, 3, "PaGeCuRs0R"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 1},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					PageCursor: []byte{1},
					Records:    []*model_cassandra.Response{{Author: "expected-username"}}},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("encrypt failure"),
				Times:        1,
			},
		}, {
			name:             "success",
			path:             "/stats-page/success/",
			query:            fmt.Sprintf(testScoresQuery["stats"], quizUUID, 3, "PaGeCuRs0R"),
			expectCursor:     "tHisIsAnEnCrYPtEdCUrS0r",
			expectPageSize:   3,
			expectNumRecords: 3,
			expectErr:        false,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 1},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					PageCursor: []byte("cursor to next page"),
					Records:    []*model_cassandra.Response{{Author: "expected-username"}, {}, {}},
					PageSize:   3,
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "tHisIsAnEnCrYPtEdCUrS0r",
				Times:        1,
			},
		}, {
			name:             "success no cursor",
			path:             "/stats-page/success-no-cursor/",
			query:            fmt.Sprintf(testScoresQuery["stats_page_size"], quizUUID, 3),
			expectPageSize:   0,
			expectNumRecords: 3,
			expectErr:        false,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 0},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					PageCursor: []byte{},
					Records:    []*model_cassandra.Response{{Author: "expected-username"}, {}, {}},
					PageSize:   3,
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:             "success quiz id only",
			path:             "/stats-page/success-quiz-id-only/",
			query:            fmt.Sprintf(testScoresQuery["stats_quiz_id"], quizUUID),
			expectCursor:     "tHisIsAnEnCrYPtEdCUrS0r",
			expectPageSize:   7,
			expectNumRecords: 3,
			expectErr:        false,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 0},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					PageCursor: []byte("cursor to next page"),
					Records:    []*model_cassandra.Response{{Author: "expected-username"}, {}, {}},
					PageSize:   7,
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "tHisIsAnEnCrYPtEdCUrS0r",
				Times:        1,
			},
		}, {
			name:      "cursor encryption failure",
			path:      "/stats-page/cursor-encryption-failure/",
			query:     fmt.Sprintf(testScoresQuery["stats_quiz_id"], quizUUID),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 0},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					PageCursor: []byte("cursor to next page"),
					Records:    []*model_cassandra.Response{{Author: "expected-username"}, {}, {}},
					PageSize:   7,
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "tHisIsAnEnCrYPtEdCUrS0r",
				OutputErr:    errors.New("encrypting cursor failed"),
				Times:        1,
			},
		}, {
			name:      "cursor decryption failure",
			path:      "/stats-page/cursor-decryption-failure/",
			query:     fmt.Sprintf(testScoresQuery["stats"], quizUUID, 3, "PaGeCuRs0R"),
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{
				OutputErr: errors.New("decrypting cursor failed"),
				Times:     1,
			},
			cassandraStatsData: &http_common.MockCassandraData{Times: 0},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
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

			gomock.InOrder(
				// Validate JWT.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.OutputParam1,
					testCase.authValidateJWTData.OutputParam2,
					testCase.authValidateJWTData.OutputErr,
				).Times(testCase.authValidateJWTData.Times),
				// Decrypt cursor page.
				mockAuth.EXPECT().DecryptFromString(gomock.Any()).Return(
					testCase.authDecryptData.OutputParam1,
					testCase.authDecryptData.OutputErr,
				).Times(testCase.authDecryptData.Times),
				// Get stats.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraStatsData.OutputParam,
					testCase.cassandraStatsData.OutputErr,
				).Times(testCase.cassandraStatsData.Times),
				// Encrypt cursor page.
				mockAuth.EXPECT().EncryptToString(gomock.Any()).Return(
					testCase.authEncryptData.OutputParam1,
					testCase.authEncryptData.OutputErr,
				).Times(testCase.authEncryptData.Times),
			)

			// Endpoint setup for test.
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrader, zapLogger))

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

				statsResponse := data.(map[string]any)["getStats"].(map[string]any)

				metadata := statsResponse["metadata"].(map[string]any)
				quizID := metadata["quizID"].(string)
				require.Equal(t, quizUUID, quizID, "quiz id did not match expected")
				numRecords := int(metadata["numRecords"].(float64))
				require.Equal(t, testCase.expectNumRecords, numRecords, "record count does not match expected")

				nextPage := statsResponse["nextPage"].(map[string]any)
				pageSize := int(nextPage["pageSize"].(float64))
				require.Equal(t, testCase.expectPageSize, pageSize, "page size does not match expected")
				cursor := nextPage["cursor"].(string)
				require.Equal(t, testCase.expectCursor, cursor, "cursor does not match expected")

			}
		})
	}
}

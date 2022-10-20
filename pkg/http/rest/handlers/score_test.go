package http_handlers

import (
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
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
)

func TestGetScore(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		authValidateJWTData *mockAuthData
		cassandraReadData   *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/score/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/score/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db read not found",
			path:           "/score/db-read-not-found/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusNotFound,
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
			name:           "success",
			path:           "/score/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.Response{
					Username:     "mock response card",
					Score:        99.99,
					QuizResponse: nil,
					QuizID:       gocql.TimeUUID(),
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
			router.GET(testCase.path+":quiz_id", GetScore(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("GET", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quizResponse id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_rest.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body")

				require.True(t, len(response.Message) != 0, "did not receive quiz response message")

				responseMap, ok := response.Payload.(map[string]any)
				require.True(t, ok, "failed to convert payload to an index-able map")
				require.NotEqual(t, 0, responseMap["score"], "failed to get score from payload")
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedLen         int
		expectedStatus      int
		authValidateJWTData *mockAuthData
		cassandraQuizData   *mockCassandraData
		cassandraStatsData  *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/stats/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraQuizData: &mockCassandraData{
				times: 0,
			},
			cassandraStatsData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/stats/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraQuizData: &mockCassandraData{
				times: 0,
			},
			cassandraStatsData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db read not found",
			path:           "/stats/db-read-not-found/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusNotFound,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraQuizData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "failed to find quiz",
					Status:  http.StatusNotFound,
				},
				times: 1,
			},
			cassandraStatsData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db username does not match author",
			path:           "/stats/db-username-does-not-match-author/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &mockAuthData{
				outputParam1: "jwt username",
				times:        1,
			},
			cassandraQuizData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{
					Author: "quiz author",
				},
				times: 1,
			},
			cassandraStatsData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db score not found",
			path:           "/stats/db-score-not-found/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusNotFound,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected username",
				times:        1,
			},
			cassandraQuizData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{
					Author: "expected username",
				},
				times: 1,
			},
			cassandraStatsData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "scorecard not found",
					Status:  http.StatusNotFound,
				},
				times: 1,
			},
		}, {
			name:           "success",
			path:           "/stats/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedLen:    3,
			expectedStatus: http.StatusOK,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected username",
				times:        1,
			},
			cassandraQuizData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{
					Author: "expected username",
				},
				times: 1,
			},
			cassandraStatsData: &mockCassandraData{
				outputParam: []*model_cassandra.Response{
					{
						Username: "username 1",
						Score:    99.9,
						QuizID:   gocql.TimeUUID(),
					}, {
						Username: "username 2",
						Score:    95.9,
						QuizID:   gocql.TimeUUID(),
					}, {
						Username: "username 3",
						Score:    90.9,
						QuizID:   gocql.TimeUUID(),
					},
				},
				times: 1,
			},
		}, {
			name:           "success no responses",
			path:           "/stats/success-no-responses/",
			quizId:         gocql.TimeUUID().String(),
			expectedLen:    0,
			expectedStatus: http.StatusOK,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected username",
				times:        1,
			},
			cassandraQuizData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{
					Author: "expected username",
				},
				times: 1,
			},
			cassandraStatsData: &mockCassandraData{
				outputParam: []*model_cassandra.Response{},
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

			gomock.InOrder(
				// Get quiz.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraQuizData.outputParam,
					testCase.cassandraQuizData.outputErr,
				).Times(testCase.cassandraQuizData.times),
				// Get stats.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraStatsData.outputParam,
					testCase.cassandraStatsData.outputErr,
				).Times(testCase.cassandraStatsData.times),
			)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

			// Endpoint setup for test.
			router.GET(testCase.path+":quiz_id", GetStats(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("GET", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quizResponse id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_rest.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body")

				require.True(t, len(response.Message) != 0, "did not receive quiz response message")

				responseList := response.Payload.([]any)
				require.Equal(t, testCase.expectedLen, len(responseList), "incorrect payload record count")
				for _, item := range responseList {
					responseMap, ok := item.(map[string]any)
					require.True(t, ok, "failed to convert payload to an index-able map")
					require.NotEqual(t, 0, responseMap["score"], "failed to get score from payload")
				}
			}
		})
	}
}

func TestPrepareStatsRequest(t *testing.T) {
	testCases := []struct {
		name       string
		pageCursor string
		pageSize   string
		quizId     gocql.UUID
		authData   struct {
			times  int
			output []byte
			err    error
		}
		expectPageSize  int
		expectErr       require.ErrorAssertionFunc
		expectNil       require.ValueAssertionFunc
		expectNilCursor require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name:       "non-numeric page size",
			pageCursor: "some page cursor string",
			pageSize:   "this should be a natural number",
			quizId:     gocql.TimeUUID(),
			authData: struct {
				times  int
				output []byte
				err    error
			}{times: 0},
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:       "failed to decrypt cursor",
			pageCursor: "some page cursor string",
			pageSize:   "3",
			quizId:     gocql.TimeUUID(),
			authData: struct {
				times  int
				output []byte
				err    error
			}{times: 1, output: nil, err: fmt.Errorf("failure decrypting")},
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:       "success - not natural number page size",
			pageCursor: "some page cursor string",
			pageSize:   "0",
			quizId:     gocql.TimeUUID(),
			authData: struct {
				times  int
				output []byte
				err    error
			}{times: 1, output: []byte{1}},
			expectPageSize:  10,
			expectErr:       require.NoError,
			expectNil:       require.NotNil,
			expectNilCursor: require.NotNil,
		}, {
			name:       "success - empty page cursor",
			pageCursor: "",
			pageSize:   "3",
			quizId:     gocql.TimeUUID(),
			authData: struct {
				times  int
				output []byte
				err    error
			}{times: 0},
			expectPageSize:  3,
			expectErr:       require.NoError,
			expectNil:       require.NotNil,
			expectNilCursor: require.Nil,
		}, {
			name:       "success",
			pageCursor: "some page cursor string",
			pageSize:   "3",
			quizId:     gocql.TimeUUID(),
			authData: struct {
				times  int
				output []byte
				err    error
			}{times: 1, output: []byte{1}},
			expectPageSize:  3,
			expectErr:       require.NoError,
			expectNil:       require.NotNil,
			expectNilCursor: require.NotNil,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)

			mockAuth.EXPECT().DecryptFromString(gomock.Any()).Return(
				testCase.authData.output,
				testCase.authData.err,
			).Times(testCase.authData.times)

			req, err := prepareStatsRequest(mockAuth, testCase.quizId, testCase.pageCursor, testCase.pageSize)
			testCase.expectErr(t, err, "error expectation condition failed")
			testCase.expectNil(t, req, "nil expectation condition failed")

			if err == nil {
				require.Equal(t, testCase.expectPageSize, req.PageSize, "expected page size check failed")
				testCase.expectNilCursor(t, req.PageCursor, "page cursor nil expectation failed")
			}

		})
	}
}

func TestGetStatsPage(t *testing.T) {
	router := getRouter()
	_ = router
}

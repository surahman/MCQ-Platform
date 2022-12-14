package http_handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	http_common "github.com/surahman/mcq-platform/pkg/http"
	"github.com/surahman/mcq-platform/pkg/mocks"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
	"github.com/surahman/mcq-platform/pkg/redis"
)

func TestCreateQuiz(t *testing.T) {
	router := http_common.GetTestRouter()

	testCases := []struct {
		name                string
		path                string
		expectedStatus      int
		quiz                *model_cassandra.QuizCore
		authValidateJWTData *http_common.MockAuthData
		cassandraCreateData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/create/empty-token",
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "empty quiz",
			path:           "/create/empty-quiz",
			expectedStatus: http.StatusBadRequest,
			quiz:           &model_cassandra.QuizCore{},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    nil,
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "valid quiz",
			path:           "/create/valid-quiz",
			expectedStatus: http.StatusOK,
			quiz:           testQuizData["myPubQuiz"].QuizCore,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    nil,
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 1,
			},
		}, {
			name:           "invalid quiz",
			path:           "/create/invalid-quiz",
			expectedStatus: http.StatusBadRequest,
			quiz:           testQuizData["invalidOptionsNoPubQuiz"].QuizCore,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    nil,
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "db failure internal",
			path:           "/create/db failure internal",
			expectedStatus: http.StatusInternalServerError,
			quiz:           testQuizData["myPubQuiz"].QuizCore,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    nil,
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				Times: 1,
			},
		}, {
			name:           "db failure conflict",
			path:           "/create/db failure conflict",
			expectedStatus: http.StatusConflict,
			quiz:           testQuizData["myPubQuiz"].QuizCore,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    nil,
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusConflict,
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

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraCreateData.OutputParam,
				testCase.cassandraCreateData.OutputErr,
			).Times(testCase.cassandraCreateData.Times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.OutputParam1,
				testCase.authValidateJWTData.OutputParam2,
				testCase.authValidateJWTData.OutputErr,
			).Times(testCase.authValidateJWTData.Times)

			quiz := testCase.quiz
			quizJson, err := json.Marshal(&quiz)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			// Endpoint setup for test.
			router.POST(testCase.path, CreateQuiz(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBuffer(quizJson))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quiz id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_http.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")

				require.Containsf(t, response.Message, "created quiz", "got incorrect message %s", response.Message)

				quizId, ok := response.Payload.(string)
				require.True(t, ok, "quiz id not returned")
				require.True(t, len(quizId) > 0, "no quiz id returned")
			}
		})
	}
}

func TestViewQuiz(t *testing.T) {
	router := http_common.GetTestRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		expectAnswers       require.BoolAssertionFunc
		authValidateJWTData *http_common.MockAuthData
		cassandraReadData   *http_common.MockCassandraData
		redisGetData        *http_common.MockRedisData
		redisSetData        *http_common.MockRedisData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/view/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/view/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			redisGetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
		}, {
			name:           "db failure",
			path:           "/view/db-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusNotFound,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{Message: "db failure", Status: http.StatusNotFound},
				Times:     1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
		}, {
			name:           "unpublished not owner",
			path:           "/view/unpublished-not-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "not owner",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: cassandra.GetTestQuizzes()["myNoPubQuiz"],
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
		}, {
			name:           "published but deleted not owner",
			path:           "/view/published-but-deleted-not-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "not owner",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: cassandra.GetTestQuizzes()["myPubQuizDeleted"],
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
		}, {
			name:           "published not owner",
			path:           "/view/published-not-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.False,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "not owner",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: cassandra.GetTestQuizzes()["myPubQuiz"],
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 1,
			},
		}, {
			name:           "published owner",
			path:           "/view/published-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "user-2",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: cassandra.GetTestQuizzes()["myPubQuiz"],
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 1,
			},
		}, {
			name:           "unpublished owner",
			path:           "/view/unpublished-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "user-3",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: cassandra.GetTestQuizzes()["myNoPubQuiz"],
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
		}, {
			name:           "published deleted owner",
			path:           "/view/published-deleted-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "user-2",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: cassandra.GetTestQuizzes()["myPubQuizDeleted"],
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
		}, {
			name:           "cache set failure",
			path:           "/view/cache-set-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "user-2",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: cassandra.GetTestQuizzes()["myPubQuiz"],
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Err: &redis.Error{
					Message: "cache set error",
					Code:    redis.ErrorCacheSet,
				},
				Times: 1,
			},
		}, {
			name:           "cache hit",
			path:           "/view/cache-hit/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "user-2",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: *testQuizData["myPubQuiz"],
				Times:  1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
			redisSetData: &http_common.MockRedisData{
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
			mockRedis := mocks.NewMockRedis(mockCtrl)

			gomock.InOrder(
				// Check JWT.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.OutputParam1,
					testCase.authValidateJWTData.OutputParam2,
					testCase.authValidateJWTData.OutputErr,
				).Times(testCase.authValidateJWTData.Times),

				// Get data from Redis.
				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).Return(
					testCase.redisGetData.Err,
				).SetArg(
					1,
					testCase.redisGetData.Param2,
				).Times(testCase.redisGetData.Times),

				// Get data from Cassandra.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraReadData.OutputParam,
					testCase.cassandraReadData.OutputErr,
				).Times(testCase.cassandraReadData.Times),

				// Set data from Redis.
				mockRedis.EXPECT().Set(gomock.Any(), gomock.Any()).Return(
					testCase.redisSetData.Err,
				).Times(testCase.redisSetData.Times),
			)

			// Endpoint setup for test.
			router.GET(testCase.path+":quiz_id", ViewQuiz(zapLogger, mockAuth, mockCassandra, mockRedis))
			req, _ := http.NewRequest("GET", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quiz id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_http.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")

				require.True(t, len(response.Message) != 0, "did not receive quiz id in response")

				questions, ok := response.Payload.(map[string]any)["questions"]
				require.True(t, ok, "failed to extract questions.")

				for _, question := range questions.([]any) {
					_, found := question.(map[string]any)["answers"]
					testCase.expectAnswers(t, found, "answer condition failed")
				}
			}
		})
	}
}

func TestDeleteQuiz(t *testing.T) {
	router := http_common.GetTestRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		authValidateJWTData *http_common.MockAuthData
		redisGetData        *http_common.MockRedisData
		redisDeleteData     *http_common.MockRedisData
		cassandraDeleteData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/delete/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Times: 0,
			},
			redisDeleteData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/delete/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			redisGetData: &http_common.MockRedisData{
				Times: 0,
			},
			redisDeleteData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "cache failure - eviction",
			path:           "/delete/cache-failure-eviction",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected username",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param1: model_cassandra.Quiz{
					Author: "expected username",
				},
				Times: 1,
			},
			redisDeleteData: &http_common.MockRedisData{
				Err: &redis.Error{
					Message: "some error not dealing with a cache miss",
					Code:    redis.ErrorUnknown,
				},
				Times: 1,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "cache failure - unauthorized",
			path:           "/delete/cache-failure-unauthorized",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected username",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param1: model_cassandra.Quiz{
					Author: "not owner",
				},
				Times: 1,
			},
			redisDeleteData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "db failure",
			path:           "/delete/db-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected username",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param1: model_cassandra.Quiz{
					Author: "expected username",
				},
				Times: 1,
			},
			redisDeleteData: &http_common.MockRedisData{
				Times: 1,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				Times: 1,
			},
		}, {
			name:           "db unauthorized",
			path:           "/delete/db- unauthorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected username",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param1: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			redisDeleteData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusForbidden,
				},
				Times: 1,
			},
		}, {
			name:           "success",
			path:           "/delete/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected username",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param1: model_cassandra.Quiz{
					Author: "expected username",
				},
				Times: 1,
			},
			redisDeleteData: &http_common.MockRedisData{
				Times: 1,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 1,
			},
		}, {
			name:           "success - cache miss",
			path:           "/delete/success-cache-miss/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected owner",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param1: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			redisDeleteData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraDeleteData: &http_common.MockCassandraData{
				Times: 1,
			},
		}, {
			name:           "cache get failure",
			path:           "/delete/cache-get-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected owner",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param1: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache failure",
					Code:    redis.ErrorUnknown,
				},
				Times: 1,
			},
			redisDeleteData: &http_common.MockRedisData{
				Times: 0,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)

			gomock.InOrder(
				// Check authorization.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.OutputParam1,
					testCase.authValidateJWTData.OutputParam2,
					testCase.authValidateJWTData.OutputErr,
				).Times(testCase.authValidateJWTData.Times),

				// Cache delete.
				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).SetArg(
					1,
					testCase.redisGetData.Param1,
				).Return(
					testCase.redisGetData.Err,
				).Times(testCase.redisGetData.Times),

				// Cache delete.
				mockRedis.EXPECT().Del(gomock.Any()).Return(
					testCase.redisDeleteData.Err,
				).Times(testCase.redisDeleteData.Times),

				// Update record in Cassandra.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraDeleteData.OutputParam,
					testCase.cassandraDeleteData.OutputErr,
				).Times(testCase.cassandraDeleteData.Times),
			)

			// Endpoint setup for test.
			router.DELETE(testCase.path+":quiz_id", DeleteQuiz(zapLogger, mockAuth, mockCassandra, mockRedis))
			req, _ := http.NewRequest("DELETE", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quiz id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_http.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")

				require.True(t, len(response.Message) != 0, "did not receive quiz id message response")
				require.True(t, len(response.Payload.(string)) != 0, "did not receive quiz id in response")
			}
		})
	}
}

func TestPublishQuiz(t *testing.T) {
	router := http_common.GetTestRouter()

	testCases := []struct {
		name                 string
		path                 string
		quizId               string
		expectedStatus       int
		authValidateJWTData  *http_common.MockAuthData
		cassandraPublishData *http_common.MockCassandraData
		cassandraGetData     *http_common.MockCassandraData
		redisSetData         *http_common.MockRedisData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/publish/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			cassandraPublishData: &http_common.MockCassandraData{
				OutputErr: nil,
				Times:     0,
			},
			cassandraGetData: &http_common.MockCassandraData{Times: 0},
			redisSetData:     &http_common.MockRedisData{Times: 0},
		}, {
			name:           "invalid quiz id",
			path:           "/publish/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			cassandraPublishData: &http_common.MockCassandraData{
				Times: 0,
			},
			cassandraGetData: &http_common.MockCassandraData{Times: 0},
			redisSetData:     &http_common.MockRedisData{Times: 0},
		}, {
			name:           "db failure",
			path:           "/publish/db-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraPublishData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				Times: 1,
			},
			cassandraGetData: &http_common.MockCassandraData{Times: 0},
			redisSetData:     &http_common.MockRedisData{Times: 0},
		}, {
			name:           "db unauthorized",
			path:           "/publish/db-unauthorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraPublishData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusForbidden,
				},
				Times: 1,
			},
			cassandraGetData: &http_common.MockCassandraData{Times: 0},
			redisSetData:     &http_common.MockRedisData{Times: 0},
		}, {
			name:           "success - db read failure",
			path:           "/publish/success-db-read-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraPublishData: &http_common.MockCassandraData{
				OutputErr: nil,
				Times:     1,
			},
			cassandraGetData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Quiz{},
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				Times: 1,
			},
			redisSetData: &http_common.MockRedisData{Times: 0},
		}, {
			name:           "success - cache set failure",
			path:           "/publish/success-cache-set-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraPublishData: &http_common.MockCassandraData{
				OutputErr: nil,
				Times:     1,
			},
			cassandraGetData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Quiz{},
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Err:   redis.NewError("something bad happened!"),
				Times: 1,
			},
		}, {
			name:           "success",
			path:           "/publish/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraPublishData: &http_common.MockCassandraData{
				OutputErr: nil,
				Times:     1,
			},
			cassandraGetData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Quiz{},
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{Times: 1},
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
			mockRedis := mocks.NewMockRedis(mockCtrl)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.OutputParam1,
				testCase.authValidateJWTData.OutputParam2,
				testCase.authValidateJWTData.OutputErr,
			).Times(testCase.authValidateJWTData.Times)

			gomock.InOrder(
				// Publish.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraPublishData.OutputParam,
					testCase.cassandraPublishData.OutputErr,
				).Times(testCase.cassandraPublishData.Times),
				// View.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraGetData.OutputParam,
					testCase.cassandraGetData.OutputErr,
				).Times(testCase.cassandraGetData.Times),
			)

			mockRedis.EXPECT().Set(gomock.Any(), gomock.Any()).Return(
				testCase.redisSetData.Err,
			).Times(testCase.redisSetData.Times)

			// Endpoint setup for test.
			router.PATCH(testCase.path+":quiz_id", PublishQuiz(zapLogger, mockAuth, mockCassandra, mockRedis))
			req, _ := http.NewRequest("PATCH", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quiz id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_http.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")

				require.True(t, len(response.Message) != 0, "did not receive quiz id message response")
				require.True(t, len(response.Payload.(string)) != 0, "did not receive quiz id in response")
			}
		})
	}
}

func TestUpdateQuiz(t *testing.T) {
	router := http_common.GetTestRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		quiz                *model_cassandra.QuizCore
		authValidateJWTData *http_common.MockAuthData
		cassandraUpdateData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/update/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			cassandraUpdateData: &http_common.MockCassandraData{
				OutputErr: nil,
				Times:     0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/update/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			cassandraUpdateData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "request validate failure",
			path:           "/update/request-validate-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusBadRequest,
			quiz:           testQuizData["invalidOptionsNoPubQuiz"].QuizCore,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraUpdateData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "db unauthorized",
			path:           "/update/db-unauthorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			quiz:           testQuizData["myNoPubQuiz"].QuizCore,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraUpdateData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusForbidden,
				},
				Times: 1,
			},
		}, {
			name:           "db failure",
			path:           "/update/db-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			quiz:           testQuizData["myNoPubQuiz"].QuizCore,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraUpdateData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				Times: 1,
			},
		}, {
			name:           "success",
			path:           "/update/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			quiz:           testQuizData["myNoPubQuiz"].QuizCore,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraUpdateData: &http_common.MockCassandraData{
				OutputErr: nil,
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
				testCase.cassandraUpdateData.OutputParam,
				testCase.cassandraUpdateData.OutputErr,
			).Times(testCase.cassandraUpdateData.Times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.OutputParam1,
				testCase.authValidateJWTData.OutputParam2,
				testCase.authValidateJWTData.OutputErr,
			).Times(testCase.authValidateJWTData.Times)

			quiz := testCase.quiz
			quizJson, err := json.Marshal(&quiz)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			// Endpoint setup for test.
			router.PATCH(testCase.path+":quiz_id", UpdateQuiz(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("PATCH", testCase.path+testCase.quizId, bytes.NewBuffer(quizJson))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quiz id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_http.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")

				require.True(t, len(response.Message) != 0, "did not receive quiz id message response")
				require.True(t, len(response.Payload.(string)) != 0, "did not receive quiz id in response")
			}
		})
	}
}

func TestTakeQuiz(t *testing.T) {
	router := http_common.GetTestRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		quizResponse        *model_cassandra.QuizResponse
		authValidateJWTData *http_common.MockAuthData
		redisGetData        *http_common.MockRedisData
		cassandraReadData   *http_common.MockCassandraData
		redisSetData        *http_common.MockRedisData
		cassandraTakeData   *http_common.MockCassandraData
		graderData          *http_common.MockGraderData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/take/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				Times: 0,
			},
			graderData: &http_common.MockGraderData{
				Times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/take/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			redisGetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				Times: 0,
			},
			graderData: &http_common.MockGraderData{
				Times: 0,
			},
		}, {
			name:           "request validate failure",
			path:           "/take/request-validate-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusBadRequest,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{-1}, {1, 2, 3, 4}}},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				Times: 0,
			},
			graderData: &http_common.MockGraderData{
				Times: 0,
			},
		}, {
			name:           "db read unauthorized",
			path:           "/take/db-read-unauthorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusForbidden,
				},
				Times: 1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				Times: 0,
			},
			graderData: &http_common.MockGraderData{
				Times: 0,
			},
		}, {
			name:           "db read failure",
			path:           "/take/db-read-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				Times: 1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				Times: 0,
			},
			graderData: &http_common.MockGraderData{
				Times: 0,
			},
		}, {
			name:           "quiz unpublished",
			path:           "/take/quiz-unpublished/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Quiz{
					IsPublished: false,
				},
				Times: 1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				Times: 0,
			},
			graderData: &http_common.MockGraderData{
				Times: 0,
			},
		}, {
			name:           "quiz deleted",
			path:           "/take/quiz-deleted/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Quiz{
					IsPublished: true,
					IsDeleted:   true,
				},
				Times: 1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				Times: 0,
			},
			graderData: &http_common.MockGraderData{
				Times: 0,
			},
		}, {
			name:           "grader failure",
			path:           "/take/grader-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusBadRequest,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Quiz{IsPublished: true},
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 1,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				Times: 0,
			},
			graderData: &http_common.MockGraderData{
				OutputErr: errors.New("grader failure"),
				Times:     1,
			},
		}, {
			name:           "db take unauthorized",
			path:           "/take/db-take-unauthorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Quiz{IsPublished: true},
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 1,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "db take auth failure",
					Status:  http.StatusForbidden,
				},
				Times: 1,
			},
			graderData: &http_common.MockGraderData{
				OutputParam: 1.333,
				Times:       1,
			},
		}, {
			name:           "db take failure",
			path:           "/take/db-take-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Quiz{IsPublished: true},
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 1,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "db take auth failure",
					Status:  http.StatusInternalServerError,
				},
				Times: 1,
			},
			graderData: &http_common.MockGraderData{
				OutputParam: 1.333,
				Times:       1,
			},
		}, {
			name:           "success",
			path:           "/take/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Quiz{IsPublished: true},
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 1,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				Times: 1,
			},
			graderData: &http_common.MockGraderData{
				OutputParam: 1.333,
				Times:       1,
			},
		}, {
			name:           "success - cache set failure",
			path:           "/take/success-cache-set-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{},
				Err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				Times: 1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Quiz{IsPublished: true},
				OutputErr:   nil,
				Times:       1,
			},
			redisSetData: &http_common.MockRedisData{
				Err: &redis.Error{
					Message: "internal error",
					Code:    redis.ErrorUnknown,
				},
				Times: 1,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				Times: 1,
			},
			graderData: &http_common.MockGraderData{
				OutputParam: 1.333,
				Times:       1,
			},
		}, {
			name:           "success - cache hit",
			path:           "/take/success-cache-hit/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			redisGetData: &http_common.MockRedisData{
				Param2: model_cassandra.Quiz{IsPublished: true},
				Times:  1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputErr: nil,
				Times:     0,
			},
			redisSetData: &http_common.MockRedisData{
				Times: 0,
			},
			cassandraTakeData: &http_common.MockCassandraData{
				Times: 1,
			},
			graderData: &http_common.MockGraderData{
				OutputParam: 1.333,
				Times:       1,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)
			mockGrader := mocks.NewMockGrading(mockCtrl)

			gomock.InOrder(
				// Validate JWT.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.OutputParam1,
					testCase.authValidateJWTData.OutputParam2,
					testCase.authValidateJWTData.OutputErr,
				).Times(testCase.authValidateJWTData.Times),

				// Cache call.
				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).SetArg(
					1,
					testCase.redisGetData.Param2,
				).Return(
					testCase.redisGetData.Err,
				).Times(testCase.redisGetData.Times),

				// Read quizResponse.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraReadData.OutputParam,
					testCase.cassandraReadData.OutputErr,
				).Times(testCase.cassandraReadData.Times),

				// Cache set.
				mockRedis.EXPECT().Set(gomock.Any(), gomock.Any()).Return(
					testCase.redisSetData.Err,
				).Times(testCase.redisSetData.Times),

				// Grade quiz.
				mockGrader.EXPECT().Grade(gomock.Any(), gomock.Any()).Return(
					testCase.graderData.OutputParam,
					testCase.graderData.OutputErr,
				).Times(testCase.graderData.Times),

				// Submit response.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraTakeData.OutputParam,
					testCase.cassandraTakeData.OutputErr,
				).Times(testCase.cassandraTakeData.Times),
			)

			responseJson, err := json.Marshal(&testCase.quizResponse)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			// Endpoint setup for test.
			router.POST(testCase.path+":quiz_id", TakeQuiz(zapLogger, mockAuth, mockCassandra, mockRedis, mockGrader))
			req, _ := http.NewRequest("POST", testCase.path+testCase.quizId, bytes.NewBuffer(responseJson))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quizResponse id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_http.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body")

				require.True(t, len(response.Message) != 0, "did not receive quiz response message")

				responseMap, ok := response.Payload.(map[string]any)
				require.True(t, ok, "failed to convert payload to an index-able map")
				require.NotEqual(t, 0, responseMap["score"], "failed to get score from payload")
			}
		})
	}
}

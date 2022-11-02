package http_handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/mocks"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
	"github.com/surahman/mcq-platform/pkg/redis"
)

func TestGetQuiz(t *testing.T) {
	testCases := []struct {
		name              string
		quiz              *model_cassandra.Quiz
		redisGetData      *mockRedisData
		cassandraReadData *mockCassandraData
		redisSetData      *mockRedisData
		expectErr         require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name: "not found",
			quiz: &model_cassandra.Quiz{QuizID: gocql.TimeUUID()},
			redisGetData: &mockRedisData{
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: nil,
				outputErr: &cassandra.Error{
					Message: "not found",
					Status:  http.StatusNotFound,
				},
				times: 1,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
			expectErr: require.Error,
		}, {
			name: "cache hit",
			quiz: &model_cassandra.Quiz{},
			redisGetData: &mockRedisData{
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
			expectErr: require.NoError,
		}, {
			name: "cache miss, db read success, cache store success",
			quiz: testQuizData["myPubQuiz"],
			redisGetData: &mockRedisData{
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testQuizData["myPubQuiz"],
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 1,
			},
			expectErr: require.NoError,
		}, {
			name: "cache miss, db read success, cache set failure",
			quiz: testQuizData["myPubQuiz"],
			redisGetData: &mockRedisData{
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testQuizData["myPubQuiz"],
				times:       1,
			},
			redisSetData: &mockRedisData{
				err: &redis.Error{
					Message: "cache failure",
					Code:    redis.ErrorUnknown,
				},
				times: 1,
			},
			expectErr: require.NoError,
		}, {
			name: "cache miss, db read success, not published",
			quiz: testQuizData["myNoPubQuiz"],
			redisGetData: &mockRedisData{
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: testQuizData["myNoPubQuiz"],
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
			expectErr: require.NoError,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockRedis := mocks.NewMockRedis(mockCtrl)
			mockCassandra := mocks.NewMockCassandra(mockCtrl)

			gomock.InOrder(
				// Cache call.
				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).Return(
					testCase.redisGetData.err,
				).Times(testCase.redisGetData.times),
				// Cassandra read.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraReadData.outputParam,
					testCase.cassandraReadData.outputErr,
				).Times(testCase.cassandraReadData.times),
				// Cache set.
				mockRedis.EXPECT().Set(gomock.Any(), gomock.Any()).Return(
					testCase.redisSetData.err,
				).Times(testCase.redisSetData.times),
			)

			actual, err := getQuiz(testCase.quiz.QuizID, mockCassandra, mockRedis)

			testCase.expectErr(t, err, "error expectation failed")
			if err != nil {
				return
			}

			require.True(t, reflect.DeepEqual(testCase.quiz, actual), "returned quiz does not match expected")
		})
	}
}

func TestCreateQuiz(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		expectedStatus      int
		quiz                *model_cassandra.QuizCore
		authValidateJWTData *mockAuthData
		cassandraCreateData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/create/empty-token",
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "empty quiz",
			path:           "/create/empty-quiz",
			expectedStatus: http.StatusBadRequest,
			quiz:           &model_cassandra.QuizCore{},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "valid quiz",
			path:           "/create/valid-quiz",
			expectedStatus: http.StatusOK,
			quiz:           testQuizData["myPubQuiz"].QuizCore,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 1,
			},
		}, {
			name:           "invalid quiz",
			path:           "/create/invalid-quiz",
			expectedStatus: http.StatusBadRequest,
			quiz:           testQuizData["invalidOptionsNoPubQuiz"].QuizCore,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db failure internal",
			path:           "/create/db failure internal",
			expectedStatus: http.StatusInternalServerError,
			quiz:           testQuizData["myPubQuiz"].QuizCore,
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
			name:           "db failure conflict",
			path:           "/create/db failure conflict",
			expectedStatus: http.StatusConflict,
			quiz:           testQuizData["myPubQuiz"].QuizCore,
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

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraCreateData.outputParam,
				testCase.cassandraCreateData.outputErr,
			).Times(testCase.cassandraCreateData.times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

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
				response := model_rest.Success{}
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
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		expectAnswers       require.BoolAssertionFunc
		authValidateJWTData *mockAuthData
		cassandraReadData   *mockCassandraData
		redisGetData        *mockRedisData
		redisSetData        *mockRedisData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/view/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			redisGetData: &mockRedisData{
				times: 0,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/view/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			redisGetData: &mockRedisData{
				times: 0,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
		}, {
			name:           "db failure",
			path:           "/view/db-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusNotFound,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputErr: &cassandra.Error{Message: "db failure", Status: http.StatusNotFound},
				times:     1,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
		}, {
			name:           "unpublished not owner",
			path:           "/view/unpublished-not-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &mockAuthData{
				outputParam1: "not owner",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myNoPubQuiz"],
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
		}, {
			name:           "published but deleted not owner",
			path:           "/view/published-but-deleted-not-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &mockAuthData{
				outputParam1: "not owner",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myPubQuizDeleted"],
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
		}, {
			name:           "published not owner",
			path:           "/view/published-not-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.False,
			authValidateJWTData: &mockAuthData{
				outputParam1: "not owner",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myPubQuiz"],
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 1,
			},
		}, {
			name:           "published owner",
			path:           "/view/published-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &mockAuthData{
				outputParam1: "user-2",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myPubQuiz"],
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 1,
			},
		}, {
			name:           "unpublished owner",
			path:           "/view/unpublished-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &mockAuthData{
				outputParam1: "user-3",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myNoPubQuiz"],
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
		}, {
			name:           "published deleted owner",
			path:           "/view/published-deleted-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &mockAuthData{
				outputParam1: "user-2",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myPubQuizDeleted"],
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
		}, {
			name:           "cache set failure",
			path:           "/view/cache-set-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &mockAuthData{
				outputParam1: "user-2",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss error",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myPubQuiz"],
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				err: &redis.Error{
					Message: "cache set error",
					Code:    redis.ErrorCacheSet,
				},
				times: 1,
			},
		}, {
			name:           "cache hit",
			path:           "/view/cache-hit/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &mockAuthData{
				outputParam1: "user-2",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: *testQuizData["myPubQuiz"],
				times:  1,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			redisSetData: &mockRedisData{
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
			mockRedis := mocks.NewMockRedis(mockCtrl)

			gomock.InOrder(
				// Check JWT.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.outputParam1,
					testCase.authValidateJWTData.outputParam2,
					testCase.authValidateJWTData.outputErr,
				).Times(testCase.authValidateJWTData.times),

				// Get data from Redis.
				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).Return(
					testCase.redisGetData.err,
				).SetArg(
					1,
					testCase.redisGetData.param2,
				).Times(testCase.redisGetData.times),

				// Get data from Cassandra.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraReadData.outputParam,
					testCase.cassandraReadData.outputErr,
				).Times(testCase.cassandraReadData.times),

				// Set data from Redis.
				mockRedis.EXPECT().Set(gomock.Any(), gomock.Any()).Return(
					testCase.redisSetData.err,
				).Times(testCase.redisSetData.times),
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
				response := model_rest.Success{}
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
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		authValidateJWTData *mockAuthData
		redisDeleteData     *mockRedisData
		cassandraDeleteData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/delete/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			redisDeleteData: &mockRedisData{
				times: 0,
			},
			cassandraDeleteData: &mockCassandraData{
				outputErr: nil,
				times:     0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/delete/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			redisDeleteData: &mockRedisData{
				times: 0,
			},
			cassandraDeleteData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "cache failure - eviction",
			path:           "/delete/cache-failure-eviction",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisDeleteData: &mockRedisData{
				err: &redis.Error{
					Message: "some error not dealing with a cache miss",
					Code:    redis.ErrorUnknown,
				},
				times: 1,
			},
			cassandraDeleteData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db failure",
			path:           "/delete/db-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisDeleteData: &mockRedisData{
				times: 1,
			},
			cassandraDeleteData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				times: 1,
			},
		}, {
			name:           "db unauthorized",
			path:           "/delete/db- unauthorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisDeleteData: &mockRedisData{
				times: 1,
			},
			cassandraDeleteData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusForbidden,
				},
				times: 1,
			},
		}, {
			name:           "success",
			path:           "/delete/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &mockAuthData{
				outputParam1: "not owner",
				times:        1,
			},
			redisDeleteData: &mockRedisData{
				times: 1,
			},
			cassandraDeleteData: &mockCassandraData{
				outputErr: nil,
				times:     1,
			},
		}, {
			name:           "success - cache miss",
			path:           "/delete/success-cache-miss/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &mockAuthData{
				outputParam1: "not owner",
				times:        1,
			},
			redisDeleteData: &mockRedisData{
				err: &redis.Error{
					Message: "unable to locate key on Redis cluster",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraDeleteData: &mockCassandraData{
				outputErr: nil,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraDeleteData.outputParam,
				testCase.cassandraDeleteData.outputErr,
			).Times(testCase.cassandraDeleteData.times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

			mockRedis.EXPECT().Del(gomock.Any()).Return(
				testCase.redisDeleteData.err,
			).Times(testCase.redisDeleteData.times)

			// Endpoint setup for test.
			router.DELETE(testCase.path+":quiz_id", DeleteQuiz(zapLogger, mockAuth, mockCassandra, mockRedis))
			req, _ := http.NewRequest("DELETE", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quiz id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_rest.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")

				require.True(t, len(response.Message) != 0, "did not receive quiz id message response")
				require.True(t, len(response.Payload.(string)) != 0, "did not receive quiz id in response")
			}
		})
	}
}

func TestPublishQuiz(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                 string
		path                 string
		quizId               string
		expectedStatus       int
		authValidateJWTData  *mockAuthData
		cassandraPublishData *mockCassandraData
		cassandraGetData     *mockCassandraData
		redisSetData         *mockRedisData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/publish/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraPublishData: &mockCassandraData{
				outputErr: nil,
				times:     0,
			},
			cassandraGetData: &mockCassandraData{times: 0},
			redisSetData:     &mockRedisData{times: 0},
		}, {
			name:           "invalid quiz id",
			path:           "/publish/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraPublishData: &mockCassandraData{
				times: 0,
			},
			cassandraGetData: &mockCassandraData{times: 0},
			redisSetData:     &mockRedisData{times: 0},
		}, {
			name:           "db failure",
			path:           "/publish/db-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraPublishData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				times: 1,
			},
			cassandraGetData: &mockCassandraData{times: 0},
			redisSetData:     &mockRedisData{times: 0},
		}, {
			name:           "db unauthorized",
			path:           "/publish/db-unauthorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraPublishData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusForbidden,
				},
				times: 1,
			},
			cassandraGetData: &mockCassandraData{times: 0},
			redisSetData:     &mockRedisData{times: 0},
		}, {
			name:           "success - db read failure",
			path:           "/publish/success-db-read-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraPublishData: &mockCassandraData{
				outputErr: nil,
				times:     1,
			},
			cassandraGetData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{},
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				times: 1,
			},
			redisSetData: &mockRedisData{times: 0},
		}, {
			name:           "success - cache set failure",
			path:           "/publish/success-cache-set-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraPublishData: &mockCassandraData{
				outputErr: nil,
				times:     1,
			},
			cassandraGetData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{},
				times:       1,
			},
			redisSetData: &mockRedisData{
				err:   redis.NewError("something bad happened!"),
				times: 1,
			},
		}, {
			name:           "success",
			path:           "/publish/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraPublishData: &mockCassandraData{
				outputErr: nil,
				times:     1,
			},
			cassandraGetData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{},
				times:       1,
			},
			redisSetData: &mockRedisData{times: 1},
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
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

			gomock.InOrder(
				// Publish.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraPublishData.outputParam,
					testCase.cassandraPublishData.outputErr,
				).Times(testCase.cassandraPublishData.times),
				// View.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraGetData.outputParam,
					testCase.cassandraGetData.outputErr,
				).Times(testCase.cassandraGetData.times),
			)

			mockRedis.EXPECT().Set(gomock.Any(), gomock.Any()).Return(
				testCase.redisSetData.err,
			).Times(testCase.redisSetData.times)

			// Endpoint setup for test.
			router.PATCH(testCase.path+":quiz_id", PublishQuiz(zapLogger, mockAuth, mockCassandra, mockRedis))
			req, _ := http.NewRequest("PATCH", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quiz id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_rest.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")

				require.True(t, len(response.Message) != 0, "did not receive quiz id message response")
				require.True(t, len(response.Payload.(string)) != 0, "did not receive quiz id in response")
			}
		})
	}
}

func TestUpdateQuiz(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		quiz                *model_cassandra.QuizCore
		authValidateJWTData *mockAuthData
		cassandraUpdateData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/update/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraUpdateData: &mockCassandraData{
				outputErr: nil,
				times:     0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/update/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraUpdateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "request validate failure",
			path:           "/update/request-validate-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusBadRequest,
			quiz:           testQuizData["invalidOptionsNoPubQuiz"].QuizCore,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraUpdateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db unauthorized",
			path:           "/update/db-unauthorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			quiz:           testQuizData["myNoPubQuiz"].QuizCore,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraUpdateData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusForbidden,
				},
				times: 1,
			},
		}, {
			name:           "db failure",
			path:           "/update/db-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			quiz:           testQuizData["myNoPubQuiz"].QuizCore,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraUpdateData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				times: 1,
			},
		}, {
			name:           "success",
			path:           "/update/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			quiz:           testQuizData["myNoPubQuiz"].QuizCore,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraUpdateData: &mockCassandraData{
				outputErr: nil,
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
				testCase.cassandraUpdateData.outputParam,
				testCase.cassandraUpdateData.outputErr,
			).Times(testCase.cassandraUpdateData.times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

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
				response := model_rest.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")

				require.True(t, len(response.Message) != 0, "did not receive quiz id message response")
				require.True(t, len(response.Payload.(string)) != 0, "did not receive quiz id in response")
			}
		})
	}
}

func TestTakeQuiz(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		quizResponse        *model_cassandra.QuizResponse
		authValidateJWTData *mockAuthData
		redisGetData        *mockRedisData
		cassandraReadData   *mockCassandraData
		redisSetData        *mockRedisData
		cassandraTakeData   *mockCassandraData
		graderData          *mockGraderData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/take/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			redisGetData: &mockRedisData{
				times: 0,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
			cassandraTakeData: &mockCassandraData{
				times: 0,
			},
			graderData: &mockGraderData{
				times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/take/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			redisGetData: &mockRedisData{
				times: 0,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
			cassandraTakeData: &mockCassandraData{
				times: 0,
			},
			graderData: &mockGraderData{
				times: 0,
			},
		}, {
			name:           "request validate failure",
			path:           "/take/request-validate-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusBadRequest,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{-1}, {1, 2, 3, 4}}},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				times: 0,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
			cassandraTakeData: &mockCassandraData{
				times: 0,
			},
			graderData: &mockGraderData{
				times: 0,
			},
		}, {
			name:           "db read unauthorized",
			path:           "/take/db-read-unauthorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusForbidden,
				},
				times: 1,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
			cassandraTakeData: &mockCassandraData{
				times: 0,
			},
			graderData: &mockGraderData{
				times: 0,
			},
		}, {
			name:           "db read failure",
			path:           "/take/db-read-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				times: 1,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
			cassandraTakeData: &mockCassandraData{
				times: 0,
			},
			graderData: &mockGraderData{
				times: 0,
			},
		}, {
			name:           "quiz unpublished",
			path:           "/take/quiz-unpublished/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{
					IsPublished: false,
				},
				times: 1,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
			cassandraTakeData: &mockCassandraData{
				times: 0,
			},
			graderData: &mockGraderData{
				times: 0,
			},
		}, {
			name:           "quiz deleted",
			path:           "/take/quiz-deleted/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{
					IsPublished: true,
					IsDeleted:   true,
				},
				times: 1,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
			cassandraTakeData: &mockCassandraData{
				times: 0,
			},
			graderData: &mockGraderData{
				times: 0,
			},
		}, {
			name:           "grader failure",
			path:           "/take/grader-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusBadRequest,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{IsPublished: true},
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 1,
			},
			cassandraTakeData: &mockCassandraData{
				times: 0,
			},
			graderData: &mockGraderData{
				outputErr: errors.New("grader failure"),
				times:     1,
			},
		}, {
			name:           "db take unauthorized",
			path:           "/take/db-take-unauthorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{IsPublished: true},
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 1,
			},
			cassandraTakeData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "db take auth failure",
					Status:  http.StatusForbidden,
				},
				times: 1,
			},
			graderData: &mockGraderData{
				outputParam: 1.333,
				times:       1,
			},
		}, {
			name:           "db take failure",
			path:           "/take/db-take-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{IsPublished: true},
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 1,
			},
			cassandraTakeData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "db take auth failure",
					Status:  http.StatusInternalServerError,
				},
				times: 1,
			},
			graderData: &mockGraderData{
				outputParam: 1.333,
				times:       1,
			},
		}, {
			name:           "success",
			path:           "/take/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{IsPublished: true},
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				times: 1,
			},
			cassandraTakeData: &mockCassandraData{
				times: 1,
			},
			graderData: &mockGraderData{
				outputParam: 1.333,
				times:       1,
			},
		}, {
			name:           "success - cache set failure",
			path:           "/take/success-cache-set-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{},
				err: &redis.Error{
					Message: "cache miss",
					Code:    redis.ErrorCacheMiss,
				},
				times: 1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.Quiz{IsPublished: true},
				outputErr:   nil,
				times:       1,
			},
			redisSetData: &mockRedisData{
				err: &redis.Error{
					Message: "internal error",
					Code:    redis.ErrorUnknown,
				},
				times: 1,
			},
			cassandraTakeData: &mockCassandraData{
				times: 1,
			},
			graderData: &mockGraderData{
				outputParam: 1.333,
				times:       1,
			},
		}, {
			name:           "success - cache hit",
			path:           "/take/success-cache-hit/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			quizResponse:   &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			redisGetData: &mockRedisData{
				param2: model_cassandra.Quiz{IsPublished: true},
				times:  1,
			},
			cassandraReadData: &mockCassandraData{
				outputErr: nil,
				times:     0,
			},
			redisSetData: &mockRedisData{
				times: 0,
			},
			cassandraTakeData: &mockCassandraData{
				times: 1,
			},
			graderData: &mockGraderData{
				outputParam: 1.333,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)
			mockGrader := mocks.NewMockGrading(mockCtrl)

			gomock.InOrder(
				// Validate JWT.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.outputParam1,
					testCase.authValidateJWTData.outputParam2,
					testCase.authValidateJWTData.outputErr,
				).Times(testCase.authValidateJWTData.times),

				// Cache call.
				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).SetArg(
					1,
					testCase.redisGetData.param2,
				).Return(
					testCase.redisGetData.err,
				).Times(testCase.redisGetData.times),

				// Read quizResponse.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraReadData.outputParam,
					testCase.cassandraReadData.outputErr,
				).Times(testCase.cassandraReadData.times),

				// Cache set.
				mockRedis.EXPECT().Set(gomock.Any(), gomock.Any()).Return(
					testCase.redisSetData.err,
				).Times(testCase.redisSetData.times),

				// Grade quiz.
				mockGrader.EXPECT().Grade(gomock.Any(), gomock.Any()).Return(
					testCase.graderData.outputParam,
					testCase.graderData.outputErr,
				).Times(testCase.graderData.times),

				// Submit response.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraTakeData.outputParam,
					testCase.cassandraTakeData.outputErr,
				).Times(testCase.cassandraTakeData.times),
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

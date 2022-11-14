package graphql_resolvers

import (
	"bytes"
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
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/redis"
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
		authValidateJWTData *http_common.MockAuthData
		cassandraCreateData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/create/empty-token",
			query:     testQuizQuery["create_valid"],
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "empty quiz",
			path:      "/create/empty-quiz",
			query:     testQuizQuery["create_empty"],
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    nil,
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "valid quiz",
			path:      "/create/valid-quiz",
			query:     testQuizQuery["create_valid"],
			expectErr: false,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    nil,
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 1,
			},
		}, {
			name:      "invalid quiz",
			path:      "/create/invalid-quiz",
			query:     testQuizQuery["create_invalid"],
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    nil,
				Times:        1,
			},
			cassandraCreateData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "db failure internal",
			path:      "/create/db failure internal",
			query:     testQuizQuery["create_valid"],
			expectErr: true,
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
			name:      "db failure conflict",
			path:      "/create/db failure conflict",
			query:     testQuizQuery["create_valid"],
			expectErr: true,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)     // Not called.
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

			gomock.InOrder(
				// Check JWT.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.OutputParam1,
					testCase.authValidateJWTData.OutputParam2,
					testCase.authValidateJWTData.OutputErr,
				).Times(testCase.authValidateJWTData.Times),

				// Store in database.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraCreateData.OutputParam,
					testCase.cassandraCreateData.OutputErr,
				).Times(testCase.cassandraCreateData.Times),
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

func TestMutationResolver_UpdateQuiz(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := getRouter()
	router.Use(GinContextToContextMiddleware())

	testCases := []struct {
		name                string
		path                string
		quizId              string
		query               string
		expectErr           bool
		authValidateJWTData *http_common.MockAuthData
		cassandraUpdateData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/update/empty-token/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["update_valid"],
			expectErr: true,
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
			name:      "invalid quiz id",
			path:      "/update/invalid-quiz-id",
			quizId:    "not a valid uuid",
			query:     testQuizQuery["update_valid"],
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			cassandraUpdateData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "request validate failure",
			path:      "/update/request-validate-failure/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["update_invalid"],
			expectErr: true,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraUpdateData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:      "db unauthorized",
			path:      "/update/db-unauthorized/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["update_valid"],
			expectErr: true,
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
			name:      "db failure",
			path:      "/update/db-failure/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["update_valid"],
			expectErr: true,
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
			name:      "success",
			path:      "/update/success/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["update_valid"],
			expectErr: false,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)     // Not called.
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

			gomock.InOrder(
				// Check authorization.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.OutputParam1,
					testCase.authValidateJWTData.OutputParam2,
					testCase.authValidateJWTData.OutputErr,
				).Times(testCase.authValidateJWTData.Times),

				// Send data to Cassandra.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraUpdateData.OutputParam,
					testCase.cassandraUpdateData.OutputErr,
				).Times(testCase.cassandraUpdateData.Times),
			)

			// Endpoint setup for test.
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrading, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(fmt.Sprintf(testCase.query, testCase.quizId)))
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
				quizID := data.(map[string]any)["updateQuiz"].(string)
				require.Equal(t, testCase.quizId, quizID, "actual and expected quid ids did not match")
			}
		})
	}
}

func TestQueryResolver_ViewQuiz(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := getRouter()
	router.Use(GinContextToContextMiddleware())

	testCases := []struct {
		name                string
		path                string
		quizId              string
		query               string
		expectErr           bool
		expectAnswers       require.ValueAssertionFunc
		authValidateJWTData *http_common.MockAuthData
		cassandraReadData   *http_common.MockCassandraData
		redisGetData        *http_common.MockRedisData
		redisSetData        *http_common.MockRedisData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/view/empty-token/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["view"],
			expectErr: true,
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
			name:      "invalid quiz id",
			path:      "/view/invalid-quiz-id",
			quizId:    "not a valid uuid",
			query:     testQuizQuery["view"],
			expectErr: true,
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
			name:      "db failure",
			path:      "/view/db-failure/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["view"],
			expectErr: true,
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
			name:      "unpublished not owner",
			path:      "/view/unpublished-not-owner/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["view"],
			expectErr: true,
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
			name:      "published but deleted not owner",
			path:      "/view/published-but-deleted-not-owner/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["view"],
			expectErr: true,
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
			name:          "published not owner",
			path:          "/view/published-not-owner/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.Nil,
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
			name:          "published owner",
			path:          "/view/published-owner/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.NotNil,
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
			name:          "unpublished owner",
			path:          "/view/unpublished-owner/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.NotNil,
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
			name:          "published deleted owner",
			path:          "/view/published-deleted-owner/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.NotNil,
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
			name:          "cache set failure",
			path:          "/view/cache-set-failure/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.NotNil,
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
			name:          "cache hit",
			path:          "/view/cache-hit/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.NotNil,
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
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

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
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrading, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(fmt.Sprintf(testCase.query, testCase.quizId)))
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
				quiz := model_cassandra.QuizCore{}
				jsonStr, err := json.Marshal(data.(map[string]any)["viewQuiz"])
				require.NoError(t, err, "failed to generate JSON string")
				require.NoError(t, json.Unmarshal(jsonStr, &quiz), "failed to unmarshall to quiz core")
				testCase.expectAnswers(t, quiz.Questions[0].Answers, "answers expectation failed")
			}
		})
	}
}

func TestMutationResolver_DeleteQuiz(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := getRouter()
	router.Use(GinContextToContextMiddleware())

	testCases := []struct {
		name                string
		path                string
		quizId              string
		query               string
		expectErr           bool
		authValidateJWTData *http_common.MockAuthData
		redisGetData        *http_common.MockRedisData
		redisDeleteData     *http_common.MockRedisData
		cassandraDeleteData *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/delete/empty-token/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["delete"],
			expectErr: true,
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
			name:      "invalid quiz id",
			path:      "/delete/invalid-quiz-id",
			quizId:    "not a valid uuid",
			query:     testQuizQuery["delete"],
			expectErr: true,
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
			name:      "cache failure - eviction",
			path:      "/delete/cache-failure-eviction",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["delete"],
			expectErr: true,
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
			name:      "cache failure - unauthorized",
			path:      "/delete/cache-failure-unauthorized",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["delete"],
			expectErr: true,
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
			name:      "db failure",
			path:      "/delete/db-failure/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["delete"],
			expectErr: true,
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
			name:      "db unauthorized",
			path:      "/delete/db-unauthorized/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["delete"],
			expectErr: true,
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
			name:      "success",
			path:      "/delete/success/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["delete"],
			expectErr: false,
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
			name:      "success - cache miss",
			path:      "/delete/success-cache-miss/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["delete"],
			expectErr: false,
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
			name:      "cache get failure",
			path:      "/delete/cache-get-failure/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["delete"],
			expectErr: true,
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
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

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
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrading, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(fmt.Sprintf(testCase.query, testCase.quizId)))
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
				quizID := data.(map[string]any)["deleteQuiz"].(string)
				require.Contains(t, quizID, testCase.quizId, "actual and expected quid ids did not match")
			}
		})
	}
}

func TestMutationResolver_PublishQuiz(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := getRouter()
	router.Use(GinContextToContextMiddleware())

	testCases := []struct {
		name                 string
		path                 string
		quizId               string
		query                string
		expectErr            bool
		authValidateJWTData  *http_common.MockAuthData
		cassandraPublishData *http_common.MockCassandraData
		cassandraGetData     *http_common.MockCassandraData
		redisSetData         *http_common.MockRedisData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/publish/empty-token/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["publish"],
			expectErr: true,
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
			name:      "invalid quiz id",
			path:      "/publish/invalid-quiz-id",
			quizId:    "not a valid uuid",
			query:     testQuizQuery["publish"],
			expectErr: true,
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
			name:      "db failure",
			path:      "/publish/db-failure/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["publish"],
			expectErr: true,
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
			name:      "db unauthorized",
			path:      "/publish/db-unauthorized/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["publish"],
			expectErr: true,
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
			name:      "success - db read failure",
			path:      "/publish/success-db-read-failure/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["publish"],
			expectErr: false,
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
			name:      "success - cache set failure",
			path:      "/publish/success-cache-set-failure/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["publish"],
			expectErr: false,
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
			name:      "success",
			path:      "/publish/success/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["publish"],
			expectErr: false,
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
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

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
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrading, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(fmt.Sprintf(testCase.query, testCase.quizId)))
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
				quizID := data.(map[string]any)["publishQuiz"].(string)
				require.Contains(t, quizID, testCase.quizId, "actual and expected quiz ids did not match")
			}
		})
	}
}

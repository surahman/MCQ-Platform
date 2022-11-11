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
	"github.com/surahman/mcq-platform/pkg/mocks"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
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
		authValidateJWTData *mockAuthData
		cassandraUpdateData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/update/empty-token/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["update_valid"],
			expectErr: true,
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
			name:      "invalid quiz id",
			path:      "/update/invalid-quiz-id",
			quizId:    "not a valid uuid",
			query:     testQuizQuery["update_valid"],
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraUpdateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "request validate failure",
			path:      "/update/request-validate-failure/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["update_invalid"],
			expectErr: true,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraUpdateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:      "db unauthorized",
			path:      "/update/db-unauthorized/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["update_valid"],
			expectErr: true,
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
			name:      "db failure",
			path:      "/update/db-failure/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["update_valid"],
			expectErr: true,
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
			name:      "success",
			path:      "/update/success/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["update_valid"],
			expectErr: false,
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
			mockRedis := mocks.NewMockRedis(mockCtrl)     // Not called.
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

			gomock.InOrder(
				// Check authorization.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.outputParam1,
					testCase.authValidateJWTData.outputParam2,
					testCase.authValidateJWTData.outputErr,
				).Times(testCase.authValidateJWTData.times),

				// Send data to Cassandra.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraUpdateData.outputParam,
					testCase.cassandraUpdateData.outputErr,
				).Times(testCase.cassandraUpdateData.times),
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
		authValidateJWTData *mockAuthData
		cassandraReadData   *mockCassandraData
		redisGetData        *mockRedisData
		redisSetData        *mockRedisData
	}{
		// ----- test cases start ----- //
		{
			name:      "empty token",
			path:      "/view/empty-token/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["view"],
			expectErr: true,
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
			name:      "invalid quiz id",
			path:      "/view/invalid-quiz-id",
			quizId:    "not a valid uuid",
			query:     testQuizQuery["view"],
			expectErr: true,
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
			name:      "db failure",
			path:      "/view/db-failure/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["view"],
			expectErr: true,
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
			name:      "unpublished not owner",
			path:      "/view/unpublished-not-owner/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["view"],
			expectErr: true,
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
			name:      "published but deleted not owner",
			path:      "/view/published-but-deleted-not-owner/",
			quizId:    gocql.TimeUUID().String(),
			query:     testQuizQuery["view"],
			expectErr: true,
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
			name:          "published not owner",
			path:          "/view/published-not-owner/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.Nil,
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
			name:          "published owner",
			path:          "/view/published-owner/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.NotNil,
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
			name:          "unpublished owner",
			path:          "/view/unpublished-owner/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.NotNil,
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
			name:          "published deleted owner",
			path:          "/view/published-deleted-owner/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.NotNil,
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
			name:          "cache set failure",
			path:          "/view/cache-set-failure/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.NotNil,
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
			name:          "cache hit",
			path:          "/view/cache-hit/",
			quizId:        gocql.TimeUUID().String(),
			query:         testQuizQuery["view"],
			expectErr:     false,
			expectAnswers: require.NotNil,
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
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

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

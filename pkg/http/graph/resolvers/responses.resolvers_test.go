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
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/redis"
)

func TestResponseResolver_QuizResponse(t *testing.T) {
	resolver := responseResolver{}

	testCases := []struct {
		name      string
		response  *model_cassandra.Response
		expectErr require.ErrorAssertionFunc
		expectNil require.ValueAssertionFunc
		expectLen int
	}{
		// ----- test cases start ----- //
		{
			name: "no responses",
			response: &model_cassandra.Response{
				Username:     "username",
				Author:       "author",
				Score:        1.11,
				QuizResponse: nil,
				QuizID:       gocql.UUID{},
			},
			expectErr: require.NoError,
			expectNil: require.Nil,
			expectLen: 0,
		}, {
			name: "some responses",
			response: &model_cassandra.Response{
				Username:     "username",
				Author:       "author",
				Score:        1.11,
				QuizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}}},
				QuizID:       gocql.UUID{},
			},
			expectErr: require.NoError,
			expectNil: require.NotNil,
			expectLen: 10,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response, err := resolver.QuizResponse(context.TODO(), testCase.response)
			testCase.expectErr(t, err, "error expectation failed")
			testCase.expectNil(t, response, "nil expectation failed")
			require.Equal(t, len(response), testCase.expectLen, "response size expectation mismatch")
		})
	}
}

func TestResponseResolver_QuizID(t *testing.T) {
	resolver := responseResolver{}

	testCases := []struct {
		name      string
		response  *model_cassandra.Response
		expectErr require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name: "no quiz id",
			response: &model_cassandra.Response{
				Username:     "username",
				Author:       "author",
				Score:        1.11,
				QuizResponse: nil,
			},
			expectErr: require.NoError,
		}, {
			name: "some quiz id",
			response: &model_cassandra.Response{
				Username:     "username",
				Author:       "author",
				Score:        1.11,
				QuizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}}},
				QuizID:       gocql.UUID{},
			},
			expectErr: require.NoError,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			quizID, err := resolver.QuizID(context.TODO(), testCase.response)
			testCase.expectErr(t, err, "error expectation failed")
			require.Equal(t, testCase.response.QuizID.String(), quizID, "quid id mismatch")
		})
	}
}

func TestMutationResolver_TakeQuiz(t *testing.T) {
	// Configure router and middleware that loads the Gin context for the resolvers.
	router := getRouter()
	router.Use(GinContextToContextMiddleware())

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectErr           bool
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
			name:         "empty token",
			path:         "/take/empty-token/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    true,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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
			name:         "invalid quiz id",
			path:         "/take/invalid-quiz-id",
			quizId:       "not a valid uuid",
			expectErr:    true,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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
			name:         "request validate failure",
			path:         "/take/request-validate-failure/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    true,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{-1}, {1, 2, 3, 4}}},
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
			name:         "db read unauthorized",
			path:         "/take/db-read-unauthorized/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    true,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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
			name:         "db read failure",
			path:         "/take/db-read-failure/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    true,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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
			name:         "quiz unpublished",
			path:         "/take/quiz-unpublished/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    true,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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
			name:         "quiz deleted",
			path:         "/take/quiz-deleted/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    true,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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
			name:         "grader failure",
			path:         "/take/grader-failure/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    true,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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
			name:         "db take unauthorized",
			path:         "/take/db-take-unauthorized/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    true,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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
			name:         "db take failure",
			path:         "/take/db-take-failure/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    true,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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
			name:         "success",
			path:         "/take/success/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    false,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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
			name:         "success - cache set failure",
			path:         "/take/success-cache-set-failure/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    false,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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
			name:         "success - cache hit",
			path:         "/take/success-cache-hit/",
			quizId:       gocql.TimeUUID().String(),
			expectErr:    false,
			quizResponse: &model_cassandra.QuizResponse{Responses: [][]int32{{}}},
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

			// Endpoint setup for test.
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrader, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(fmt.Sprintf(testQuizQuery["take"], testCase.quizId, testCase.quizResponse.Responses)))
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
				gradingResponse := model_cassandra.Response{}
				jsonStr, err := json.Marshal(data.(map[string]any)["takeQuiz"])
				require.NoError(t, err, "failed to generate JSON string")
				require.NoError(t, json.Unmarshal(jsonStr, &gradingResponse), "failed to unmarshall to quiz response")
				require.InDelta(t, testCase.graderData.outputParam, gradingResponse.Score, 0.01, "incorrect score received")
			}
		})
	}
}

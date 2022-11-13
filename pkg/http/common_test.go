package http

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/mocks"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
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

			actual, err := GetQuiz(testCase.quiz.QuizID, mockCassandra, mockRedis)

			testCase.expectErr(t, err, "error expectation failed")
			if err != nil {
				return
			}

			require.True(t, reflect.DeepEqual(testCase.quiz, actual), "returned quiz does not match expected")
		})
	}
}

func TestPrepareStatsRequest(t *testing.T) {
	testCases := []struct {
		name            string
		pageCursor      string
		pageSize        string
		quizId          gocql.UUID
		mockAuthData    *mockAuthData
		expectPageSize  int
		expectErr       require.ErrorAssertionFunc
		expectNil       require.ValueAssertionFunc
		expectNilCursor require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name:         "non-numeric page size",
			pageCursor:   "some page cursor string",
			pageSize:     "this should be a natural number",
			quizId:       gocql.TimeUUID(),
			mockAuthData: &mockAuthData{times: 0},
			expectErr:    require.Error,
			expectNil:    require.Nil,
		}, {
			name:       "failed to decrypt cursor",
			pageCursor: "some page cursor string",
			pageSize:   "3",
			quizId:     gocql.TimeUUID(),
			mockAuthData: &mockAuthData{
				times:        1,
				outputParam1: nil,
				outputErr:    fmt.Errorf("failure decrypting"),
			},
			expectErr: require.Error,
			expectNil: require.Nil,
		}, {
			name:       "success - not natural number page size",
			pageCursor: "some page cursor string",
			pageSize:   "0",
			quizId:     gocql.TimeUUID(),
			mockAuthData: &mockAuthData{
				times:        1,
				outputParam1: []byte{1},
			},
			expectPageSize:  10,
			expectErr:       require.NoError,
			expectNil:       require.NotNil,
			expectNilCursor: require.NotNil,
		}, {
			name:            "success - empty page cursor",
			pageCursor:      "",
			pageSize:        "3",
			quizId:          gocql.TimeUUID(),
			mockAuthData:    &mockAuthData{times: 0},
			expectPageSize:  3,
			expectErr:       require.NoError,
			expectNil:       require.NotNil,
			expectNilCursor: require.Nil,
		}, {
			name:       "success",
			pageCursor: "some page cursor string",
			pageSize:   "3",
			quizId:     gocql.TimeUUID(),
			mockAuthData: &mockAuthData{
				times:        1,
				outputParam1: []byte{1},
			},
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
				testCase.mockAuthData.outputParam1,
				testCase.mockAuthData.outputErr,
			).Times(testCase.mockAuthData.times)

			req, err := PrepareStatsRequest(mockAuth, testCase.quizId, testCase.pageCursor, testCase.pageSize)
			testCase.expectErr(t, err, "error expectation condition failed")
			testCase.expectNil(t, req, "nil expectation condition failed")

			if err == nil {
				require.Equal(t, testCase.expectPageSize, req.PageSize, "expected page size check failed")
				testCase.expectNilCursor(t, req.PageCursor, "page cursor nil expectation failed")
			}
		})
	}
}

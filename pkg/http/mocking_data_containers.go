package http

import (
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
)

// MockAuthData is the parameter data for Auth mocking that is used in the test grid.
type MockAuthData struct {
	InputParam1  string
	InputParam2  string
	OutputParam1 any
	OutputParam2 int64
	OutputErr    error
	Times        int
}

// MockCassandraData is the parameter data for Cassandra mocking that is used in the test grid.
type MockCassandraData struct {
	InputFunc   func(cassandra.Cassandra, any) (any, error)
	InputParam  any
	OutputParam any
	OutputErr   error
	Times       int
}

// MockGraderData is the parameter data for Grader mocking that is used in the test grid.
type MockGraderData struct {
	InputQuizResp *model_cassandra.QuizResponse
	InputQuiz     *model_cassandra.Quiz
	OutputParam   float64
	OutputErr     error
	Times         int
}

// MockRedisData is the parameter data for Redis mocking that is used in the test grid.
type MockRedisData struct {
	Param1 any
	Param2 any
	Err    error
	Times  int
}

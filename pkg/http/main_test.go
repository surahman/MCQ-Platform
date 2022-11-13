package http

import "github.com/surahman/mcq-platform/pkg/cassandra"

// testQuizData is the test quiz data.
var testQuizData = cassandra.GetTestQuizzes()

// mockAuthData is the parameter data for Auth mocking that is used in the test grid.
type mockAuthData struct {
	inputParam1  string
	inputParam2  string
	outputParam1 any
	outputParam2 int64
	outputErr    error
	times        int
}

// mockCassandraData is the parameter data for Cassandra mocking that is used in the test grid.
type mockCassandraData struct {
	inputFunc   func(cassandra.Cassandra, any) (any, error)
	inputParam  any
	outputParam any
	outputErr   error
	times       int
}

// mockRedisData is the parameter data for Redis mocking that is used in the test grid.
type mockRedisData struct {
	param1 any
	param2 any
	err    error
	times  int
}

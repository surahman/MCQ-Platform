package graphql_resolvers

import (
	"context"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
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

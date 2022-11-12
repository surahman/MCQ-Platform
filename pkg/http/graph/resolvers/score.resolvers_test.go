package graphql_resolvers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/model/http"
)

func TestMetadataResolver_QuizID(t *testing.T) {
	resolver := metadataResolver{}
	testCases := []struct {
		name        string
		metadata    *model_http.Metadata
		expectErr   require.ErrorAssertionFunc
		expectedLen int
	}{
		// ----- test cases start ----- //
		{
			name:        "nil metadata",
			metadata:    nil,
			expectErr:   require.Error,
			expectedLen: 0,
		}, {
			name:        "some metadata",
			metadata:    &model_http.Metadata{},
			expectErr:   require.NoError,
			expectedLen: 36,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			quizID, err := resolver.QuizID(context.TODO(), testCase.metadata)
			testCase.expectErr(t, err, "error expectation failed")
			require.Equal(t, testCase.expectedLen, len(quizID), "UUID not returned")
		})
	}
}

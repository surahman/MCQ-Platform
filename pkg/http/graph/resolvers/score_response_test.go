package graphql_resolvers

import (
	"testing"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	http_common "github.com/surahman/mcq-platform/pkg/http"
	"github.com/surahman/mcq-platform/pkg/mocks"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
)

func TestQueryResolver_prepareStatsResponse(t *testing.T) {
	encryptedCursor := "encrypted-page-cursor-byte-string"

	testCases := []struct {
		name             string
		quizId           gocql.UUID
		expectedCursor   []byte
		expectedPageSize int
		dbResponse       *model_cassandra.StatsResponse
		mockAuthData     *http_common.MockAuthData
		expectErr        require.ErrorAssertionFunc
		expectNil        require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name:             "nil cursor",
			quizId:           gocql.TimeUUID(),
			expectedCursor:   []byte{},
			expectedPageSize: 0,
			dbResponse: &model_cassandra.StatsResponse{
				PageCursor: nil,
				Records:    nil,
				PageSize:   0,
			},
			mockAuthData: &http_common.MockAuthData{Times: 0, OutputParam1: ""},
			expectErr:    require.NoError,
			expectNil:    require.NotNil,
		}, {
			name:             "empty cursor",
			quizId:           gocql.TimeUUID(),
			expectedCursor:   []byte{},
			expectedPageSize: 0,
			dbResponse: &model_cassandra.StatsResponse{
				PageCursor: []byte{},
				Records:    nil,
				PageSize:   0,
			},
			mockAuthData: &http_common.MockAuthData{Times: 0, OutputParam1: ""},
			expectErr:    require.NoError,
			expectNil:    require.NotNil,
		}, {
			name:             "cursor only",
			quizId:           gocql.TimeUUID(),
			expectedCursor:   []byte(encryptedCursor),
			expectedPageSize: 0,
			dbResponse: &model_cassandra.StatsResponse{
				PageCursor: []byte("page-cursor-byte-string"),
				Records:    nil,
				PageSize:   0,
			},
			mockAuthData: &http_common.MockAuthData{Times: 1, OutputParam1: encryptedCursor},
			expectErr:    require.NoError,
			expectNil:    require.NotNil,
		}, {
			name:             "single page only",
			quizId:           gocql.TimeUUID(),
			expectedCursor:   []byte{},
			expectedPageSize: 0,
			dbResponse: &model_cassandra.StatsResponse{
				PageCursor: nil,
				Records:    nil,
				PageSize:   1,
			},
			mockAuthData: &http_common.MockAuthData{Times: 0, OutputParam1: ""},
			expectErr:    require.NoError,
			expectNil:    require.NotNil,
		}, {
			name:             "cursor and page",
			quizId:           gocql.TimeUUID(),
			expectedCursor:   []byte(encryptedCursor),
			expectedPageSize: 3,
			dbResponse: &model_cassandra.StatsResponse{
				PageCursor: []byte("page-cursor-byte-string"),
				Records:    nil,
				PageSize:   3,
			},
			mockAuthData: &http_common.MockAuthData{Times: 1, OutputParam1: encryptedCursor},
			expectErr:    require.NoError,
			expectNil:    require.NotNil,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)

			mockAuth.EXPECT().EncryptToString(gomock.Any()).Return(
				testCase.mockAuthData.OutputParam1,
				testCase.mockAuthData.OutputErr,
			).Times(testCase.mockAuthData.Times)

			req, err := prepareStatsResponse(mockAuth, testCase.dbResponse, testCase.quizId)
			testCase.expectErr(t, err, "error expectation condition failed")
			testCase.expectNil(t, req, "nil expectation condition failed")

			if err == nil {
				require.Equal(t, testCase.expectedCursor, []byte(req.Cursor), "page cursor mismatch")
				require.Equal(t, testCase.expectedPageSize, req.PageSize, "page size mismatch")
			}
		})
	}
}

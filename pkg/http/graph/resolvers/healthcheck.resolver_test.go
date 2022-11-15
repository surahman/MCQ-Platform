package graphql_resolvers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	http_common "github.com/surahman/mcq-platform/pkg/http"
	"github.com/surahman/mcq-platform/pkg/mocks"
)

func TestQueryResolver_Healthcheck(t *testing.T) {
	router := http_common.GetTestRouter()

	query := getHealthcheckQuery()

	testCases := []struct {
		name                string
		path                string
		expectedMsg         string
		expectErr           bool
		cassandraHealthData *http_common.MockCassandraData
		redisHealthData     *http_common.MockRedisData
	}{
		// ----- test cases start ----- //
		{
			name:        "cassandra failure",
			path:        "/healthcheck/cassandra-failure",
			expectedMsg: "Cassandra",
			expectErr:   true,
			cassandraHealthData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "Cassandra failure",
					Status:  http.StatusInternalServerError,
				},
				Times: 1,
			},
			redisHealthData: &http_common.MockRedisData{Times: 0},
		}, {
			name:                "redis failure",
			path:                "/healthcheck/redis-failure",
			expectedMsg:         "Redis",
			expectErr:           true,
			cassandraHealthData: &http_common.MockCassandraData{Times: 1},
			redisHealthData: &http_common.MockRedisData{
				Err:   errors.New("Redis failure"),
				Times: 1,
			},
		}, {
			name:                "success",
			path:                "/healthcheck/success",
			expectedMsg:         "healthy",
			expectErr:           false,
			cassandraHealthData: &http_common.MockCassandraData{Times: 1},
			redisHealthData:     &http_common.MockRedisData{Times: 1},
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl) // Not called.
			mockCassandra := mocks.NewMockCassandra(mockCtrl)
			mockRedis := mocks.NewMockRedis(mockCtrl)
			mockGrading := mocks.NewMockGrading(mockCtrl) // Not called.

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraHealthData.OutputParam,
				testCase.cassandraHealthData.OutputErr,
			).Times(testCase.cassandraHealthData.Times)

			mockRedis.EXPECT().Healthcheck().Return(
				testCase.redisHealthData.Err,
			).Times(testCase.redisHealthData.Times)

			// Endpoint setup for test.
			router.POST(testCase.path, QueryHandler(testAuthHeaderKey, mockAuth, mockRedis, mockCassandra, mockGrading, zapLogger))

			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBufferString(query))
			req.Header.Set("Content-Type", "application/json")
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
				// Auth token is expected.
				data, ok := response["data"]
				require.True(t, ok, "data key expected but not set")
				require.Equal(t, "OK", data.(map[string]any)["healthcheck"].(string), "healthcheck did not return OK status")
			}
		})
	}
}

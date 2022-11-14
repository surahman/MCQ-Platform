package http_handlers

import (
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
	model_rest "github.com/surahman/mcq-platform/pkg/model/http"
)

func TestHealthcheck(t *testing.T) {
	router := http_common.GetTestRouter()

	testCases := []struct {
		name                string
		path                string
		expectedMsg         string
		expectedStatus      int
		cassandraHealthData *http_common.MockCassandraData
		redisHealthData     *http_common.MockRedisData
	}{
		// ----- test cases start ----- //
		{
			name:           "cassandra failure",
			path:           "/healthcheck/cassandra-failure",
			expectedMsg:    "Cassandra",
			expectedStatus: http.StatusServiceUnavailable,
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
			expectedStatus:      http.StatusServiceUnavailable,
			cassandraHealthData: &http_common.MockCassandraData{Times: 1},
			redisHealthData: &http_common.MockRedisData{
				Err:   errors.New("Redis failure"),
				Times: 1,
			},
		}, {
			name:                "success",
			path:                "/healthcheck/success",
			expectedMsg:         "healthy",
			expectedStatus:      http.StatusOK,
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
			mockCassandra := mocks.NewMockCassandra(mockCtrl)
			mockRedis := mocks.NewMockRedis(mockCtrl)

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraHealthData.OutputParam,
				testCase.cassandraHealthData.OutputErr,
			).Times(testCase.cassandraHealthData.Times)

			mockRedis.EXPECT().Healthcheck().Return(
				testCase.redisHealthData.Err,
			).Times(testCase.redisHealthData.Times)

			// Endpoint setup for test.
			router.GET(testCase.path, Healthcheck(zapLogger, mockCassandra, mockRedis))
			req, _ := http.NewRequest("GET", testCase.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			response := model_rest.Success{}
			require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")
			require.Containsf(t, response.Message, testCase.expectedMsg, "got incorrect message %s", response.Message)
		})
	}
}

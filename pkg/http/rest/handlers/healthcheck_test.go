package http_handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	"github.com/surahman/mcq-platform/pkg/mocks"
	model_rest "github.com/surahman/mcq-platform/pkg/model/http"
)

func TestHealthcheck(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		expectedMsg         string
		expectedStatus      int
		cassandraHealthData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "cassandra failure",
			path:           "/healthcheck/cassandra-failure",
			expectedMsg:    "Cassandra",
			expectedStatus: http.StatusServiceUnavailable,
			cassandraHealthData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "Cassandra failure",
					Status:  http.StatusInternalServerError,
				},
				times: 1,
			},
		}, {
			name:           "success",
			path:           "/healthcheck/success",
			expectedMsg:    "healthy",
			expectedStatus: http.StatusOK,
			cassandraHealthData: &mockCassandraData{
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
			mockCassandra := mocks.NewMockCassandra(mockCtrl)

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraHealthData.outputParam,
				testCase.cassandraHealthData.outputErr,
			).Times(testCase.cassandraHealthData.times)

			// Endpoint setup for test.
			router.GET(testCase.path, Healthcheck(zapLogger, mockCassandra))
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

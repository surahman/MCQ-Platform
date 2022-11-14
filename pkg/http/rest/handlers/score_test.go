package http_handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
	http_common "github.com/surahman/mcq-platform/pkg/http"
	"github.com/surahman/mcq-platform/pkg/mocks"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
	"github.com/surahman/mcq-platform/pkg/model/http"
)

func TestGetScore(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		authValidateJWTData *http_common.MockAuthData
		cassandraReadData   *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/score/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/score/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			cassandraReadData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "db read not found",
			path:           "/score/db-read-not-found/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusNotFound,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusNotFound,
				},
				Times: 1,
			},
		}, {
			name:           "success",
			path:           "/score/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        1,
			},
			cassandraReadData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.Response{
					Username:     "mock response card",
					Score:        99.99,
					QuizResponse: nil,
					QuizID:       gocql.TimeUUID(),
				},
				Times: 1,
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

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraReadData.OutputParam,
				testCase.cassandraReadData.OutputErr,
			).Times(testCase.cassandraReadData.Times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.OutputParam1,
				testCase.authValidateJWTData.OutputParam2,
				testCase.authValidateJWTData.OutputErr,
			).Times(testCase.authValidateJWTData.Times)

			// Endpoint setup for test.
			router.GET(testCase.path+":quiz_id", GetScore(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("GET", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quizResponse id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_http.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body")

				require.True(t, len(response.Message) != 0, "did not receive quiz response message")

				responseMap, ok := response.Payload.(map[string]any)
				require.True(t, ok, "failed to convert payload to an index-able map")
				require.NotEqual(t, 0, responseMap["score"], "failed to get score from payload")
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedLen         int
		expectedStatus      int
		authValidateJWTData *http_common.MockAuthData
		cassandraStatsData  *http_common.MockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/stats/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			cassandraStatsData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/stats/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			cassandraStatsData: &http_common.MockCassandraData{
				Times: 0,
			},
		}, {
			name:           "db score not found",
			path:           "/stats/db-score-not-found/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusNotFound,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected username",
				Times:        1,
			},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Message: "scorecard not found",
					Status:  http.StatusNotFound,
				},
				Times: 1,
			},
		}, {
			name:           "success",
			path:           "/stats/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedLen:    3,
			expectedStatus: http.StatusOK,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected username",
				Times:        1,
			},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: []*model_cassandra.Response{
					{
						Username: "username 1",
						Author:   "expected username",
						Score:    99.9,
						QuizID:   gocql.TimeUUID(),
					}, {
						Username: "username 2",
						Author:   "expected username",
						Score:    95.9,
						QuizID:   gocql.TimeUUID(),
					}, {
						Username: "username 3",
						Author:   "expected username",
						Score:    90.9,
						QuizID:   gocql.TimeUUID(),
					},
				},
				Times: 1,
			},
		}, {
			name:           "not authorized",
			path:           "/stats/not-authorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedLen:    3,
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected username",
				Times:        1,
			},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: []*model_cassandra.Response{
					{
						Username: "username 1",
						Author:   "not the author",
						Score:    99.9,
						QuizID:   gocql.TimeUUID(),
					}, {
						Username: "username 2",
						Author:   "not the author",
						Score:    95.9,
						QuizID:   gocql.TimeUUID(),
					}, {
						Username: "username 3",
						Author:   "not the author",
						Score:    90.9,
						QuizID:   gocql.TimeUUID(),
					},
				},
				Times: 1,
			},
		}, {
			name:           "success no responses",
			path:           "/stats/success-no-responses/",
			quizId:         gocql.TimeUUID().String(),
			expectedLen:    0,
			expectedStatus: http.StatusNotFound,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected username",
				Times:        1,
			},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: []*model_cassandra.Response{},
				Times:       1,
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

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraStatsData.OutputParam,
				testCase.cassandraStatsData.OutputErr,
			).Times(testCase.cassandraStatsData.Times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.OutputParam1,
				testCase.authValidateJWTData.OutputParam2,
				testCase.authValidateJWTData.OutputErr,
			).Times(testCase.authValidateJWTData.Times)

			// Endpoint setup for test.
			router.GET(testCase.path+":quiz_id", GetStats(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("GET", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quizResponse id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_http.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body")

				require.True(t, len(response.Message) != 0, "did not receive quiz response message")

				responseList := response.Payload.([]any)
				require.Equal(t, testCase.expectedLen, len(responseList), "incorrect payload record count")
				for _, item := range responseList {
					responseMap, ok := item.(map[string]any)
					require.True(t, ok, "failed to convert payload to an index-able map")
					require.NotEqual(t, 0, responseMap["score"], "failed to get score from payload")
				}
			}
		})
	}
}

func TestPrepareStatsResponse(t *testing.T) {
	testCases := []struct {
		name            string
		quizId          gocql.UUID
		dbResponse      *model_cassandra.StatsResponse
		mockAuthData    *http_common.MockAuthData
		expectErr       require.ErrorAssertionFunc
		expectNil       require.ValueAssertionFunc
		expectEmptyLink require.BoolAssertionFunc
		expectCursor    require.ComparisonAssertionFunc
		expectPage      require.ComparisonAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name:   "nil cursor",
			quizId: gocql.TimeUUID(),
			dbResponse: &model_cassandra.StatsResponse{
				PageCursor: nil,
				Records:    nil,
				PageSize:   0,
			},
			mockAuthData:    &http_common.MockAuthData{Times: 0, OutputParam1: ""},
			expectErr:       require.NoError,
			expectNil:       require.NotNil,
			expectEmptyLink: require.True,
			expectCursor:    require.NotContains,
			expectPage:      require.NotContains,
		}, {
			name:   "empty cursor",
			quizId: gocql.TimeUUID(),
			dbResponse: &model_cassandra.StatsResponse{
				PageCursor: []byte{},
				Records:    nil,
				PageSize:   0,
			},
			mockAuthData:    &http_common.MockAuthData{Times: 0, OutputParam1: ""},
			expectErr:       require.NoError,
			expectNil:       require.NotNil,
			expectEmptyLink: require.True,
			expectCursor:    require.NotContains,
			expectPage:      require.NotContains,
		}, {
			name:   "cursor only",
			quizId: gocql.TimeUUID(),
			dbResponse: &model_cassandra.StatsResponse{
				PageCursor: []byte("page-cursor-byte-string"),
				Records:    nil,
				PageSize:   0,
			},
			mockAuthData:    &http_common.MockAuthData{Times: 1, OutputParam1: ""},
			expectErr:       require.NoError,
			expectNil:       require.NotNil,
			expectEmptyLink: require.False,
			expectCursor:    require.Contains,
			expectPage:      require.NotContains,
		}, {
			name:   "page only",
			quizId: gocql.TimeUUID(),
			dbResponse: &model_cassandra.StatsResponse{
				PageCursor: nil,
				Records:    nil,
				PageSize:   1,
			},
			mockAuthData:    &http_common.MockAuthData{Times: 0, OutputParam1: ""},
			expectErr:       require.NoError,
			expectNil:       require.NotNil,
			expectEmptyLink: require.True,
			expectCursor:    require.NotContains,
			expectPage:      require.NotContains,
		}, {
			name:   "cursor and page",
			quizId: gocql.TimeUUID(),
			dbResponse: &model_cassandra.StatsResponse{
				PageCursor: []byte("page-cursor-byte-string"),
				Records:    nil,
				PageSize:   3,
			},
			mockAuthData:    &http_common.MockAuthData{Times: 1, OutputParam1: ""},
			expectErr:       require.NoError,
			expectNil:       require.NotNil,
			expectEmptyLink: require.False,
			expectCursor:    require.Contains,
			expectPage:      require.Contains,
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
				testCase.expectEmptyLink(t, len(req.Links.NextPage) == 0, "link existence condition failed")
				testCase.expectCursor(t, req.Links.NextPage, "?pageCursor=", "page cursor condition failed")
				testCase.expectPage(t, req.Links.NextPage, "&pageSize=", "page size condition failed")
			}
		})
	}
}

func TestGetStatsPage(t *testing.T) {
	router := getRouter()
	testCases := []struct {
		name                string
		path                string
		quizId              string
		querySegment        string
		expectedLen         int
		expectedStatus      int
		expectLink          require.BoolAssertionFunc
		authValidateJWTData *http_common.MockAuthData
		authDecryptData     *http_common.MockAuthData
		cassandraStatsData  *http_common.MockCassandraData
		authEncryptData     *http_common.MockAuthData
	}{
		// ----- test cases start ----- //
		{
			name:           "bad uuid",
			path:           "/stats-page/bad-uuid/",
			quizId:         "face palm",
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusBadRequest,
			expectLink:     require.False,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
			authDecryptData: &http_common.MockAuthData{Times: 0},
			cassandraStatsData: &http_common.MockCassandraData{
				Times: 0,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:           "empty token",
			path:           "/stats-page/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusInternalServerError,
			expectLink:     require.False,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("invalid token"),
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 0},
			cassandraStatsData: &http_common.MockCassandraData{
				Times: 0,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:           "db read invalid user",
			path:           "/stats-page/failed-db-read-invalid-user/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusForbidden,
			expectLink:     require.False,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 1},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					Records: []*model_cassandra.Response{{Author: "UNexpected-username"}},
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:           "db read no records",
			path:           "/stats-page/failed-db-read-no-records/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusNotFound,
			expectLink:     require.False,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 1},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					Records: []*model_cassandra.Response{},
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:           "db quiz read valid user invalid page size",
			path:           "/stats-page/failed-db-quiz-read-valid-user-invalid-page-size/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=ThisShouldBeANaturalNumber",
			expectedLen:    0,
			expectedStatus: http.StatusBadRequest,
			expectLink:     require.False,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 0},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					Records: []*model_cassandra.Response{{Author: "expected-username"}}},
				Times: 0,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:           "db stat read failure",
			path:           "/stats-page/db-stat-read-failure/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusInternalServerError,
			expectLink:     require.False,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 1},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputErr: &cassandra.Error{
					Status: http.StatusInternalServerError,
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
			},
		}, {
			name:           "prepare response failure",
			path:           "/stats-page/prepare-response-failure/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusInternalServerError,
			expectLink:     require.False,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 1},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					PageCursor: []byte{1},
					Records:    []*model_cassandra.Response{{Author: "expected-username"}}},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				OutputErr:    errors.New("encrypt failure"),
				Times:        1,
			},
		}, {
			name:           "success",
			path:           "/stats-page/success/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    3,
			expectedStatus: http.StatusOK,
			expectLink:     require.True,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 1},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					PageCursor: []byte("cursor to next page"),
					Records:    []*model_cassandra.Response{{Author: "expected-username"}, {}, {}},
					PageSize:   3,
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "tHisIsAnEnCrYPtEdCUrS0r",
				Times:        1,
			},
		}, {
			name:           "success no cursor",
			path:           "/stats-page/success-no-cursor/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageSize=3",
			expectedLen:    3,
			expectedStatus: http.StatusOK,
			expectLink:     require.False,
			authValidateJWTData: &http_common.MockAuthData{
				OutputParam1: "expected-username",
				Times:        1,
			},
			authDecryptData: &http_common.MockAuthData{Times: 0},
			cassandraStatsData: &http_common.MockCassandraData{
				OutputParam: &model_cassandra.StatsResponse{
					PageCursor: []byte{},
					Records:    []*model_cassandra.Response{{Author: "expected-username"}, {}, {}},
					PageSize:   3,
				},
				Times: 1,
			},
			authEncryptData: &http_common.MockAuthData{
				OutputParam1: "",
				Times:        0,
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

			gomock.InOrder(
				// Validate JWT.
				mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
					testCase.authValidateJWTData.OutputParam1,
					testCase.authValidateJWTData.OutputParam2,
					testCase.authValidateJWTData.OutputErr,
				).Times(testCase.authValidateJWTData.Times),
				// Decrypt cursor page.
				mockAuth.EXPECT().DecryptFromString(gomock.Any()).Return(
					testCase.authDecryptData.OutputParam1,
					testCase.authDecryptData.OutputErr,
				).Times(testCase.authDecryptData.Times),
				// Get stats.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraStatsData.OutputParam,
					testCase.cassandraStatsData.OutputErr,
				).Times(testCase.cassandraStatsData.Times),
				// Encrypt cursor page.
				mockAuth.EXPECT().EncryptToString(gomock.Any()).Return(
					testCase.authEncryptData.OutputParam1,
					testCase.authEncryptData.OutputErr,
				).Times(testCase.authEncryptData.Times),
			)

			// Endpoint setup for test.
			router.GET(testCase.path+":quiz_id", GetStatsPage(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("GET", testCase.path+testCase.quizId+testCase.querySegment, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify response code.
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quizResponse id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_http.StatsResponse{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body")

				require.Equal(t, testCase.expectedLen, len(response.Records), "records count does not match expected")
				require.Equal(t, testCase.expectedLen, response.Metadata.NumRecords, "metadata record count does not match expected")
				require.Equal(t, testCase.quizId, response.Metadata.QuizID.String(), "quiz id does not match expected")
				testCase.expectLink(t, len(response.Links.NextPage) != 0, "link to next page expectation failed")
			}
		})
	}
}

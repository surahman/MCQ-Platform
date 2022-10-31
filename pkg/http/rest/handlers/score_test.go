package http_handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocql/gocql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/cassandra"
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
		authValidateJWTData *mockAuthData
		cassandraReadData   *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/score/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/score/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraReadData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db read not found",
			path:           "/score/db-read-not-found/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusNotFound,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusNotFound,
				},
				times: 1,
			},
		}, {
			name:           "success",
			path:           "/score/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraReadData: &mockCassandraData{
				outputParam: &model_cassandra.Response{
					Username:     "mock response card",
					Score:        99.99,
					QuizResponse: nil,
					QuizID:       gocql.TimeUUID(),
				},
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
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockCassandra := mocks.NewMockCassandra(mockCtrl)

			mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
				testCase.cassandraReadData.outputParam,
				testCase.cassandraReadData.outputErr,
			).Times(testCase.cassandraReadData.times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

			// Endpoint setup for test.
			router.GET(testCase.path+":quiz_id", GetScore(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("GET", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quizResponse id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_rest.Success{}
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
		authValidateJWTData *mockAuthData
		cassandraStatsData  *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/stats/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraStatsData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/stats/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraStatsData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db score not found",
			path:           "/stats/db-score-not-found/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusNotFound,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected username",
				times:        1,
			},
			cassandraStatsData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "scorecard not found",
					Status:  http.StatusNotFound,
				},
				times: 1,
			},
		}, {
			name:           "success",
			path:           "/stats/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedLen:    3,
			expectedStatus: http.StatusOK,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected username",
				times:        1,
			},
			cassandraStatsData: &mockCassandraData{
				outputParam: []*model_cassandra.Response{
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
				times: 1,
			},
		}, {
			name:           "not authorized",
			path:           "/stats/not-authorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedLen:    3,
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected username",
				times:        1,
			},
			cassandraStatsData: &mockCassandraData{
				outputParam: []*model_cassandra.Response{
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
				times: 1,
			},
		}, {
			name:           "success no responses",
			path:           "/stats/success-no-responses/",
			quizId:         gocql.TimeUUID().String(),
			expectedLen:    0,
			expectedStatus: http.StatusNotFound,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected username",
				times:        1,
			},
			cassandraStatsData: &mockCassandraData{
				outputParam: []*model_cassandra.Response{},
				times:       1,
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
				testCase.cassandraStatsData.outputParam,
				testCase.cassandraStatsData.outputErr,
			).Times(testCase.cassandraStatsData.times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

			// Endpoint setup for test.
			router.GET(testCase.path+":quiz_id", GetStats(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("GET", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quizResponse id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_rest.Success{}
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

			req, err := prepareStatsRequest(mockAuth, testCase.quizId, testCase.pageCursor, testCase.pageSize)
			testCase.expectErr(t, err, "error expectation condition failed")
			testCase.expectNil(t, req, "nil expectation condition failed")

			if err == nil {
				require.Equal(t, testCase.expectPageSize, req.PageSize, "expected page size check failed")
				testCase.expectNilCursor(t, req.PageCursor, "page cursor nil expectation failed")
			}

		})
	}
}

func TestPrepareStatsResponse(t *testing.T) {
	testCases := []struct {
		name            string
		quizId          gocql.UUID
		dbResponse      *model_cassandra.StatsResponse
		mockAuthData    *mockAuthData
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
			mockAuthData:    &mockAuthData{times: 0, outputParam1: ""},
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
			mockAuthData:    &mockAuthData{times: 0, outputParam1: ""},
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
			mockAuthData:    &mockAuthData{times: 1, outputParam1: ""},
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
			mockAuthData:    &mockAuthData{times: 0, outputParam1: ""},
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
			mockAuthData:    &mockAuthData{times: 1, outputParam1: ""},
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
				testCase.mockAuthData.outputParam1,
				testCase.mockAuthData.outputErr,
			).Times(testCase.mockAuthData.times)

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
		authValidateJWTData *mockAuthData
		authDecryptData     *mockAuthData
		cassandraStatsData  *mockCassandraData
		authEncryptData     *mockAuthData
	}{
		// ----- test cases start ----- //
		{
			name:           "bad uuid",
			path:           "/stats-page/bad-uuid/",
			quizId:         "pace palm",
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusBadRequest,
			expectLink:     require.False,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			authDecryptData: &mockAuthData{times: 0},
			cassandraStatsData: &mockCassandraData{
				times: 0,
			},
			authEncryptData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
		}, {
			name:           "empty token",
			path:           "/stats-page/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusInternalServerError,
			expectLink:     require.False,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			authDecryptData: &mockAuthData{times: 0},
			cassandraStatsData: &mockCassandraData{
				times: 0,
			},
			authEncryptData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
		}, {
			name:           "db read invalid user",
			path:           "/stats-page/failed-db-read-invalid-user/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusForbidden,
			expectLink:     require.False,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected-username",
				times:        1,
			},
			authDecryptData: &mockAuthData{times: 1},
			cassandraStatsData: &mockCassandraData{
				outputParam: &model_cassandra.StatsResponse{
					Records: []*model_cassandra.Response{{Author: "UNexpected-username"}},
				},
				times: 1,
			},
			authEncryptData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
		}, {
			name:           "db read no records",
			path:           "/stats-page/failed-db-read-no-records/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusNotFound,
			expectLink:     require.False,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected-username",
				times:        1,
			},
			authDecryptData: &mockAuthData{times: 1},
			cassandraStatsData: &mockCassandraData{
				outputParam: &model_cassandra.StatsResponse{
					Records: []*model_cassandra.Response{},
				},
				times: 1,
			},
			authEncryptData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
		}, {
			name:           "db quiz read valid user invalid page size",
			path:           "/stats-page/failed-db-quiz-read-valid-user-invalid-page-size/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=ThisShouldBeANaturalNumber",
			expectedLen:    0,
			expectedStatus: http.StatusBadRequest,
			expectLink:     require.False,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected-username",
				times:        1,
			},
			authDecryptData: &mockAuthData{times: 0},
			cassandraStatsData: &mockCassandraData{
				outputParam: &model_cassandra.StatsResponse{
					Records: []*model_cassandra.Response{{Author: "expected-username"}}},
				times: 0,
			},
			authEncryptData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
		}, {
			name:           "db stat read failure",
			path:           "/stats-page/db-stat-read-failure/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusInternalServerError,
			expectLink:     require.False,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected-username",
				times:        1,
			},
			authDecryptData: &mockAuthData{times: 1},
			cassandraStatsData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Status: http.StatusInternalServerError,
				},
				times: 1,
			},
			authEncryptData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
		}, {
			name:           "prepare response failure",
			path:           "/stats-page/prepare-response-failure/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    0,
			expectedStatus: http.StatusInternalServerError,
			expectLink:     require.False,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected-username",
				times:        1,
			},
			authDecryptData: &mockAuthData{times: 1},
			cassandraStatsData: &mockCassandraData{
				outputParam: &model_cassandra.StatsResponse{
					PageCursor: []byte{1},
					Records:    []*model_cassandra.Response{{Author: "expected-username"}}},
				times: 1,
			},
			authEncryptData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("encrypt failure"),
				times:        1,
			},
		}, {
			name:           "success",
			path:           "/stats-page/success/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageCursor=PaGeCuRs0R==&pageSize=3",
			expectedLen:    3,
			expectedStatus: http.StatusOK,
			expectLink:     require.True,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected-username",
				times:        1,
			},
			authDecryptData: &mockAuthData{times: 1},
			cassandraStatsData: &mockCassandraData{
				outputParam: &model_cassandra.StatsResponse{
					PageCursor: []byte("cursor to next page"),
					Records:    []*model_cassandra.Response{{Author: "expected-username"}, {}, {}},
					PageSize:   3,
				},
				times: 1,
			},
			authEncryptData: &mockAuthData{
				outputParam1: "tHisIsAnEnCrYPtEdCUrS0r",
				times:        1,
			},
		}, {
			name:           "success no cursor",
			path:           "/stats-page/success-no-cursor/",
			quizId:         gocql.TimeUUID().String(),
			querySegment:   "?pageSize=3",
			expectedLen:    3,
			expectedStatus: http.StatusOK,
			expectLink:     require.False,
			authValidateJWTData: &mockAuthData{
				outputParam1: "expected-username",
				times:        1,
			},
			authDecryptData: &mockAuthData{times: 0},
			cassandraStatsData: &mockCassandraData{
				outputParam: &model_cassandra.StatsResponse{
					PageCursor: []byte{},
					Records:    []*model_cassandra.Response{{Author: "expected-username"}, {}, {}},
					PageSize:   3,
				},
				times: 1,
			},
			authEncryptData: &mockAuthData{
				outputParam1: "",
				times:        0,
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
					testCase.authValidateJWTData.outputParam1,
					testCase.authValidateJWTData.outputParam2,
					testCase.authValidateJWTData.outputErr,
				).Times(testCase.authValidateJWTData.times),
				// Decrypt cursor page.
				mockAuth.EXPECT().DecryptFromString(gomock.Any()).Return(
					testCase.authDecryptData.outputParam1,
					testCase.authDecryptData.outputErr,
				).Times(testCase.authDecryptData.times),
				// Get stats.
				mockCassandra.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(
					testCase.cassandraStatsData.outputParam,
					testCase.cassandraStatsData.outputErr,
				).Times(testCase.cassandraStatsData.times),
				// Encrypt cursor page.
				mockAuth.EXPECT().EncryptToString(gomock.Any()).Return(
					testCase.authEncryptData.outputParam1,
					testCase.authEncryptData.outputErr,
				).Times(testCase.authEncryptData.times),
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
				response := model_rest.StatsResponse{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body")

				require.Equal(t, testCase.expectedLen, len(response.Records), "records count does not match expected")
				require.Equal(t, testCase.expectedLen, response.Metadata.NumRecords, "metadata record count does not match expected")
				require.Equal(t, testCase.quizId, response.Metadata.QuizID.String(), "quiz id does not match expected")
				testCase.expectLink(t, len(response.Links.NextPage) != 0, "link to next page expectation failed")
			}
		})
	}
}

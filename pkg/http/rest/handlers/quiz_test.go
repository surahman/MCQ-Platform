package http_handlers

import (
	"bytes"
	"encoding/json"
	"errors"
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

func TestCreateQuiz(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		expectedStatus      int
		quiz                *model_cassandra.QuizCore
		authValidateJWTData *mockAuthData
		cassandraCreateData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/create/empty-token",
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "empty quiz",
			path:           "/create/empty-quiz",
			expectedStatus: http.StatusBadRequest,
			quiz:           &model_cassandra.QuizCore{},
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "valid quiz",
			path:           "/create/valid-quiz",
			expectedStatus: http.StatusOK,
			quiz:           testQuizData["myPubQuiz"].QuizCore,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 1,
			},
		}, {
			name:           "invalid quiz",
			path:           "/create/invalid-quiz",
			expectedStatus: http.StatusBadRequest,
			quiz:           testQuizData["invalidOptionsNoPubQuiz"].QuizCore,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db failure internal",
			path:           "/create/db failure internal",
			expectedStatus: http.StatusInternalServerError,
			quiz:           testQuizData["myPubQuiz"].QuizCore,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				times: 1,
			},
		}, {
			name:           "db failure conflict",
			path:           "/create/db failure conflict",
			expectedStatus: http.StatusConflict,
			quiz:           testQuizData["myPubQuiz"].QuizCore,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    nil,
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusConflict,
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
				testCase.cassandraCreateData.outputParam,
				testCase.cassandraCreateData.outputErr,
			).Times(testCase.cassandraCreateData.times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

			quiz := testCase.quiz
			quizJson, err := json.Marshal(&quiz)
			require.NoErrorf(t, err, "failed to marshall JSON: %v", err)

			// Endpoint setup for test.
			router.POST(testCase.path, CreateQuiz(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("POST", testCase.path, bytes.NewBuffer(quizJson))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quiz id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_rest.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")

				require.Containsf(t, response.Message, "created quiz", "got incorrect message %s", response.Message)

				quizId, ok := response.Payload.(string)
				require.True(t, ok, "quiz id not returned")
				require.True(t, len(quizId) > 0, "no quiz id returned")
			}
		})
	}
}

func TestViewQuiz(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		expectAnswers       require.BoolAssertionFunc
		quiz                *model_cassandra.QuizCore
		authValidateJWTData *mockAuthData
		cassandraCreateData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/view/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/view/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraCreateData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db failure",
			path:           "/view/db-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusNotFound,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputErr: &cassandra.Error{Message: "db failure", Status: http.StatusNotFound},
				times:     1,
			},
		}, {
			name:           "unpublished not owner",
			path:           "/view/unpublished-not-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &mockAuthData{
				outputParam1: "not owner",
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myNoPubQuiz"],
				outputErr:   nil,
				times:       1,
			},
		}, {
			name:           "published but deleted not owner",
			path:           "/view/published-but-deleted-not-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &mockAuthData{
				outputParam1: "not owner",
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myPubQuizDeleted"],
				outputErr:   nil,
				times:       1,
			},
		}, {
			name:           "published not owner",
			path:           "/view/published-not-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.False,
			authValidateJWTData: &mockAuthData{
				outputParam1: "not owner",
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myPubQuiz"],
				outputErr:   nil,
				times:       1,
			},
		}, {
			name:           "published owner",
			path:           "/view/published-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &mockAuthData{
				outputParam1: "user-2",
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myPubQuiz"],
				outputErr:   nil,
				times:       1,
			},
		}, {
			name:           "unpublished owner",
			path:           "/view/unpublished-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &mockAuthData{
				outputParam1: "user-3",
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myNoPubQuiz"],
				outputErr:   nil,
				times:       1,
			},
		}, {
			name:           "published deleted owner",
			path:           "/view/published-deleted-owner/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			expectAnswers:  require.True,
			authValidateJWTData: &mockAuthData{
				outputParam1: "user-2",
				times:        1,
			},
			cassandraCreateData: &mockCassandraData{
				outputParam: cassandra.GetTestQuizzes()["myPubQuizDeleted"],
				outputErr:   nil,
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
				testCase.cassandraCreateData.outputParam,
				testCase.cassandraCreateData.outputErr,
			).Times(testCase.cassandraCreateData.times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

			// Endpoint setup for test.
			router.GET(testCase.path+":quiz_id", ViewQuiz(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("GET", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quiz id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_rest.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")

				require.True(t, len(response.Message) != 0, "did not receive quiz id in response")

				questions, ok := response.Payload.(map[string]any)["questions"]
				require.True(t, ok, "failed to extract questions.")

				for _, question := range questions.([]any) {
					_, found := question.(map[string]any)["answers"]
					testCase.expectAnswers(t, found, "answer condition failed")
				}
			}
		})
	}
}

func TestDeleteQuiz(t *testing.T) {
	router := getRouter()

	testCases := []struct {
		name                string
		path                string
		quizId              string
		expectedStatus      int
		expectAnswers       require.BoolAssertionFunc
		quiz                *model_cassandra.QuizCore
		authValidateJWTData *mockAuthData
		cassandraDeleteData *mockCassandraData
	}{
		// ----- test cases start ----- //
		{
			name:           "empty token",
			path:           "/delete/empty-token/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				outputErr:    errors.New("invalid token"),
				times:        1,
			},
			cassandraDeleteData: &mockCassandraData{
				outputErr: nil,
				times:     0,
			},
		}, {
			name:           "invalid quiz id",
			path:           "/delete/invalid-quiz-id",
			quizId:         "not a valid uuid",
			expectedStatus: http.StatusBadRequest,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        0,
			},
			cassandraDeleteData: &mockCassandraData{
				times: 0,
			},
		}, {
			name:           "db failure",
			path:           "/delete/db-failure/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusInternalServerError,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraDeleteData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusInternalServerError,
				},
				times: 1,
			},
		}, {
			name:           "db unauthorized",
			path:           "/delete/db- unauthorized/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusForbidden,
			authValidateJWTData: &mockAuthData{
				outputParam1: "",
				times:        1,
			},
			cassandraDeleteData: &mockCassandraData{
				outputErr: &cassandra.Error{
					Message: "",
					Status:  http.StatusForbidden,
				},
				times: 1,
			},
		}, {
			name:           "success",
			path:           "/delete/success/",
			quizId:         gocql.TimeUUID().String(),
			expectedStatus: http.StatusOK,
			authValidateJWTData: &mockAuthData{
				outputParam1: "not owner",
				times:        1,
			},
			cassandraDeleteData: &mockCassandraData{
				outputErr: nil,
				times:     1,
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
				testCase.cassandraDeleteData.outputParam,
				testCase.cassandraDeleteData.outputErr,
			).Times(testCase.cassandraDeleteData.times)

			mockAuth.EXPECT().ValidateJWT(gomock.Any()).Return(
				testCase.authValidateJWTData.outputParam1,
				testCase.authValidateJWTData.outputParam2,
				testCase.authValidateJWTData.outputErr,
			).Times(testCase.authValidateJWTData.times)

			// Endpoint setup for test.
			router.DELETE(testCase.path+":quiz_id", DeleteQuiz(zapLogger, mockAuth, mockCassandra))
			req, _ := http.NewRequest("DELETE", testCase.path+testCase.quizId, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify responses
			require.Equal(t, testCase.expectedStatus, w.Code, "expected status codes do not match")

			// Check message and quiz id.
			if testCase.expectedStatus == http.StatusOK {
				response := model_rest.Success{}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&response), "failed to unmarshall response body.")

				require.True(t, len(response.Message) != 0, "did not receive quiz id message response")
				require.True(t, len(response.Payload.(string)) != 0, "did not receive quiz id in response")
			}
		})
	}
}

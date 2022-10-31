package cassandra

import (
	"fmt"
	"math"
	"net/http"
	"reflect"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
)

var testUserRecords = GetTestUsers()
var testQuizRecords = GetTestQuizzes()
var testResponseRecords = GetTestResponses()

func insertTestUsers(t *testing.T) {
	_, err := truncateTableQuery(connection.db, "users")
	require.NoErrorf(t, err, "failed to truncate user table before populating")

	for _, user := range testUserRecords {
		_, err := CreateUserQuery(connection.db, user)
		require.NoErrorf(t, err, "failed to create user %v with error %v", user, err)
	}
}

func insertTestQuizzes(t *testing.T) {
	_, err := truncateTableQuery(connection.db, "quizzes")
	require.NoErrorf(t, err, "failed to truncate quizzes table before populating")

	for _, quiz := range testQuizRecords {
		_, err := CreateQuizQuery(connection.db, quiz)
		require.NoErrorf(t, err, "failed to create quiz %v with error %v", quiz, err)
	}
}

func insertTestResponses(t *testing.T) {
	_, err := truncateTableQuery(connection.db, "responses")
	require.NoErrorf(t, err, "failed to truncate responses table before populating")

	for _, response := range testResponseRecords {
		_, err := CreateResponseQuery(connection.db, response)
		require.NoErrorf(t, err, "failed to create response %v with error %v", response, err)
	}
}

func TestCreateUserQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new users.
	insertTestUsers(t)

	// Username and account id collisions.
	for key, testCase := range testUserRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			_, err := connection.db.Execute(CreateUserQuery, testCase)
			require.Error(t, err)
			require.Equal(t, http.StatusConflict, err.(*Error).Status)
		})
	}

	// New user with different username and account but duplicated fields.
	userPass := &model_cassandra.User{
		UserAccount: &model_cassandra.UserAccount{
			UserLoginCredentials: model_cassandra.UserLoginCredentials{
				Username: "user-5",
				Password: "user-pwd-1",
			},
			FirstName: "firstname-1",
			LastName:  "lastname-1",
			Email:     "user1@email-address.com",
		},
		AccountID: blake2b256("user-5"),
		IsDeleted: false,
	}
	_, err := connection.db.Execute(CreateUserQuery, userPass)
	require.NoError(t, err, "user account with non-duplicate key fields should be created.")
}

func TestDeleteUserQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new users.
	insertTestUsers(t)

	// Non-existent user.
	userPass := "user-5"
	_, err := connection.db.Execute(DeleteUserQuery, userPass)
	require.Error(t, err, "user account that does not exist")
	require.Equal(t, http.StatusNotFound, err.(*Error).Status)

	// User accounts deleted.
	for key, testCase := range testUserRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			_, err := connection.db.Execute(DeleteUserQuery, testCase.Username)
			require.NoError(t, err)
		})
	}
}

func TestReadUserQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new users.
	insertTestUsers(t)

	// Non-existent user.
	userPass := "user-5"
	_, err := connection.db.Execute(ReadUserQuery, userPass)
	require.Error(t, err, "user account that does not exist")
	require.Equal(t, http.StatusNotFound, err.(*Error).Status)

	// Check created accounts exist.
	for key, testCase := range testUserRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			resp, err := connection.db.Execute(ReadUserQuery, testCase.UserAccount.UserLoginCredentials.Username)
			require.NoError(t, err)
			actual := resp.(*model_cassandra.User)
			require.Truef(t, reflect.DeepEqual(testCase, actual), "expected user, %v, does not match actual, %v", testCase, actual)
		})
	}
}

func TestCreateQuizQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new quizzes.
	insertTestQuizzes(t)

	// Quiz id collisions.
	for key, testCase := range testQuizRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			_, err := connection.db.Execute(CreateQuizQuery, testCase)
			require.Error(t, err)
		})
	}
}

func TestReadQuizQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new users.
	insertTestQuizzes(t)

	// Non-existent quiz.
	_, err := connection.db.Execute(ReadQuizQuery, gocql.TimeUUID())
	require.Error(t, err, "quiz that does not exist")

	// Check created quizzes exist.
	for key, testCase := range testQuizRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			resp, err := connection.db.Execute(ReadQuizQuery, testCase.QuizID)
			require.NoError(t, err)
			actual := resp.(*model_cassandra.Quiz)
			require.Truef(t, reflect.DeepEqual(testCase, actual), "expected quiz, %v, does not match actual, %v", testCase, actual)
		})
	}
}

func TestUpdateQuizQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new users.
	insertTestQuizzes(t)

	var err error

	// Non-existent quiz.
	_, err = connection.db.Execute(UpdateQuizQuery, &model_cassandra.QuizMutateRequest{
		Username: "not the author",
		QuizID:   gocql.TimeUUID(),
		Quiz: &model_cassandra.Quiz{
			Author:      "someone or another",
			QuizCore:    &model_cassandra.QuizCore{},
			QuizID:      gocql.TimeUUID(),
			IsPublished: false,
			IsDeleted:   false,
		},
	})
	require.Error(t, err, "quiz that does not exist")

	expectedQuizzes := GetTestQuizzes()
	for key, testCase := range expectedQuizzes {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			testCase.Title = "updated title"
			testCase.MarkingType = "updated marking type"
			testCase.Questions[0].Description = "updated quiz description"

			req := &model_cassandra.QuizMutateRequest{
				Username: testCase.Author,
				QuizID:   testCase.QuizID,
				Quiz:     testCase,
			}

			_, err = connection.db.Execute(UpdateQuizQuery, req)
			if testCase.IsPublished || testCase.IsDeleted {
				require.Error(t, err, "update to a published record should failed")
				return
			}
			require.NoError(t, err, "update record failed")

			var resp any
			resp, err = connection.db.Execute(ReadQuizQuery, testCase.QuizID)
			require.NoError(t, err, "read record failed")

			actual := resp.(*model_cassandra.Quiz)
			require.Truef(t, reflect.DeepEqual(*testCase, *actual), "expected quiz, %v, does not match actual, %v", testCase, actual)
		})
	}
}

func TestDeleteQuizQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new users.
	insertTestQuizzes(t)

	// Non-existent quiz.
	_, errNonExistent := connection.db.Execute(DeleteQuizQuery, &model_cassandra.QuizMutateRequest{
		Username: "",
		QuizID:   gocql.TimeUUID(),
	})
	require.Error(t, errNonExistent, "quiz that does not exist")

	// Not owner deletion failures.
	for key, testCase := range testQuizRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			req := model_cassandra.QuizMutateRequest{
				Username: testCase.Author + "no-owner",
				QuizID:   testCase.QuizID,
			}
			_, err := connection.db.Execute(DeleteQuizQuery, &req)
			require.Error(t, err, "delete record succeeded with not author")
		})
	}

	// Owner deletion success.
	for key, testCase := range testQuizRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			req := model_cassandra.QuizMutateRequest{
				Username: testCase.Author,
				QuizID:   testCase.QuizID,
			}
			_, err := connection.db.Execute(DeleteQuizQuery, &req)
			require.NoError(t, err, "delete record failed")

			var resp any
			resp, err = connection.db.Execute(ReadQuizQuery, testCase.QuizID)
			require.NoError(t, err, "read quiz record failed")
			actual := resp.(*model_cassandra.Quiz)
			require.Truef(t, actual.IsDeleted, "expected quiz to be deleted but actual, %v", actual)
		})
	}
}

func TestPublishQuizQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new users.
	insertTestQuizzes(t)

	// Non-existent quiz.
	_, errNonExistent := connection.db.Execute(PublishQuizQuery, &model_cassandra.QuizMutateRequest{
		Username: "",
		QuizID:   gocql.TimeUUID(),
	})
	require.Error(t, errNonExistent, "quiz that does not exist")

	// Not owner publish failures.
	for key, testCase := range testQuizRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			req := model_cassandra.QuizMutateRequest{
				Username: testCase.Author + "no-owner",
				QuizID:   testCase.QuizID,
			}
			_, err := connection.db.Execute(PublishQuizQuery, &req)
			require.Error(t, err, "publish record succeeded with not author")
		})
	}

	// Owner publish success.
	for key, testCase := range testQuizRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			req := model_cassandra.QuizMutateRequest{
				Username: testCase.Author,
				QuizID:   testCase.QuizID,
			}
			_, err := connection.db.Execute(PublishQuizQuery, &req)
			if testCase.IsDeleted {
				require.Error(t, err, "a deleted record should not be published")
				return
			}
			require.NoError(t, err, "publish record failed")

			var resp any
			resp, err = connection.db.Execute(ReadQuizQuery, testCase.QuizID)
			require.NoError(t, err, "read quiz record failed")
			actual := resp.(*model_cassandra.Quiz)
			require.Truef(t, actual.IsPublished, "expected quiz to be published but actual, %v", actual)
		})
	}
}

func TestCreateResponseQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new responses.
	insertTestResponses(t)

	// Username and quiz id collisions.
	for key, testCase := range testResponseRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			_, err := connection.db.Execute(CreateResponseQuery, testCase)
			require.Error(t, err)
		})
	}
}

func TestReadResponseQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new responses.
	insertTestResponses(t)

	// Non-existent Response.
	nonExistentResponse := &model_cassandra.QuizMutateRequest{
		Username: "user-1",
		QuizID:   gocql.TimeUUID(),
	}
	_, err := connection.db.Execute(ReadResponseQuery, nonExistentResponse)
	require.Error(t, err, "user response that does not exist")

	// Username and quiz id collisions.
	for key, testCase := range testResponseRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			request := &model_cassandra.QuizMutateRequest{
				Username: testCase.Username,
				QuizID:   testCase.QuizID,
			}
			resp, err := connection.db.Execute(ReadResponseQuery, request)
			require.NoError(t, err)
			actual := resp.(*model_cassandra.Response)
			require.Truef(t, reflect.DeepEqual(testCase, actual), "expected response, %v, does not match actual, %v", testCase, actual)
		})
	}
}

func TestReadResponseStatisticsQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new responses.
	insertTestResponses(t)

	testCases := []struct {
		name         string
		uuid         gocql.UUID
		expectedSize int
		expectNil    require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			"Not found",
			gocql.TimeUUID(),
			0,
			require.Nil,
		}, {
			"myPubQuiz",
			quizzesUUIDMapping["myPubQuiz"],
			5,
			require.NotNil,
		}, {
			"providedPubQuiz",
			quizzesUUIDMapping["providedPubQuiz"],
			2,
			require.NotNil,
		}, {
			"providedNoPubQuiz",
			quizzesUUIDMapping["providedNoPubQuiz"],
			0,
			require.Nil,
		}, {
			"myNoPubQuiz",
			quizzesUUIDMapping["myNoPubQuiz"],
			0,
			require.Nil,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			responseSlice, _ := connection.db.Execute(ReadResponseStatisticsQuery, testCase.uuid)

			actual := responseSlice.([]*model_cassandra.Response)
			testCase.expectNil(t, actual, "returned array does not meet nil expectation")
			require.Equal(t, testCase.expectedSize, len(actual), "length of the response slices do not match")
		})
	}
}

func TestHealthcheckQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	response, err := HealthcheckQuery(connection.db, nil)
	require.NoError(t, err, "healthcheck query returned an error")
	require.NotNil(t, response, "response to healthcheck query is nil")
	require.Truef(t, len(response.(string)) > 0, "release version string empty")
}

func TestReadResponseStatisticsPageQuery(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new responses.
	insertTestResponses(t)

	testCases := []struct {
		name                string
		expectedRecordCount int
		request             *model_cassandra.StatsRequest
		expectNil           require.ValueAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name:                "not found",
			expectedRecordCount: 0,
			request: &model_cassandra.StatsRequest{
				QuizID:     gocql.UUID{},
				PageCursor: nil,
				PageSize:   0,
			},
			expectNil: require.Nil,
		}, {
			name:                "pages of size 5",
			expectedRecordCount: 5,
			request: &model_cassandra.StatsRequest{
				QuizID:     quizzesUUIDMapping["myPubQuiz"],
				PageCursor: nil,
				PageSize:   5,
			},
		}, {
			name:                "pages of size 2",
			expectedRecordCount: 5,
			request: &model_cassandra.StatsRequest{
				QuizID:     quizzesUUIDMapping["myPubQuiz"],
				PageCursor: nil,
				PageSize:   2,
			},
		}, {
			name:                "pages of size 1",
			expectedRecordCount: 5,
			request: &model_cassandra.StatsRequest{
				QuizID:     quizzesUUIDMapping["myPubQuiz"],
				PageCursor: nil,
				PageSize:   1,
			},
		}, {
			name:                "pages of size 6",
			expectedRecordCount: 5,
			request: &model_cassandra.StatsRequest{
				QuizID:     quizzesUUIDMapping["myPubQuiz"],
				PageCursor: nil,
				PageSize:   8,
			},
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var actual *model_cassandra.StatsResponse

			if testCase.expectedRecordCount < 1 {
				response, err := connection.db.Execute(ReadResponseStatisticsPageQuery, testCase.request)
				require.NoError(t, err, "failed to execute paged response statistics query")
				actual = response.(*model_cassandra.StatsResponse)
				testCase.expectNil(t, actual.Records, "returned records array does not meet nil expectation")
				require.Equal(t, testCase.expectedRecordCount, len(actual.Records), "number of response records doesn't match expected")
				return
			}

			// Iterate over all pages and tally up the total records retrieved.
			expectedPageCount := int(math.Ceil(float64(testCase.expectedRecordCount / testCase.request.PageSize)))
			pageCount := 0
			recordCount := 0
			request := &model_cassandra.StatsRequest{
				QuizID:     testCase.request.QuizID,
				PageCursor: nil,
				PageSize:   testCase.request.PageSize,
			}
			for {
				response, err := connection.db.Execute(ReadResponseStatisticsPageQuery, request)
				require.NoError(t, err, "failed to execute paged response statistics query")
				actual = response.(*model_cassandra.StatsResponse)
				require.Equal(t, testCase.request.PageSize, actual.PageSize, "page size not set in response")
				recordCount += len(actual.Records)

				if len(actual.PageCursor) == 0 {
					break
				}
				pageCount++
				request.PageCursor = actual.PageCursor
			}
			require.Equal(t, expectedPageCount, pageCount, "incorrect number of pages received")
			require.Equal(t, testCase.expectedRecordCount, recordCount, "number of response records doesn't match expected")
		})
	}
}

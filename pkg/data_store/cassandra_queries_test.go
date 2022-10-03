package data_store

import (
	"fmt"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
)

var testUserRecords = GetTestUsers()
var testResponseRecords = GetTestResponses()

func insertTestUsers(t *testing.T) {
	_, err := truncateTableQuery(connection.db, "users")
	require.NoErrorf(t, err, "failed to truncate user table before populating")

	for _, user := range testUserRecords {
		_, err := CreateUserQuery(connection.db, user)
		require.NoErrorf(t, err, "failed to create user %v with error %v", user, err)
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
		})
	}

	// New user with different username and account but duplicated fields.
	userPass := &model_cassandra.User{
		UserAccount: &model_cassandra.UserAccount{
			Username:  "user-5",
			Password:  "user-pwd-1",
			FirstName: "firstname-1",
			LastName:  "lastname-1",
			Email:     "user1@email-address.com",
		},
		AccountID: "user-account-id-5",
		IsDeleted: false,
	}
	_, err := connection.db.Execute(CreateUserQuery, userPass)
	require.NoError(t, err, "user account with non-duplicate key fields should be created.")
}

func TestDeleteUserQuery(t *testing.T) {
	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new users.
	insertTestUsers(t)

	// Non-existent user.
	userPass := &model_cassandra.User{
		UserAccount: &model_cassandra.UserAccount{
			Username:  "user-5",
			Password:  "user-pwd-1",
			FirstName: "firstname-1",
			LastName:  "lastname-1",
			Email:     "user1@email-address.com",
		},
		AccountID: "user-account-id-5",
		IsDeleted: false,
	}
	_, err := connection.db.Execute(DeleteUserQuery, userPass)
	require.Error(t, err, "user account that does not exist")

	// Username and account id collisions.
	for key, testCase := range testUserRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			_, err := connection.db.Execute(DeleteUserQuery, testCase)
			require.NoError(t, err)
		})
	}
}

func TestReadUserQuery(t *testing.T) {
	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new users.
	insertTestUsers(t)

	// Non-existent user.
	userPass := &model_cassandra.User{
		UserAccount: &model_cassandra.UserAccount{
			Username:  "user-5",
			Password:  "user-pwd-1",
			FirstName: "firstname-1",
			LastName:  "lastname-1",
			Email:     "user1@email-address.com",
		},
		AccountID: "user-account-id-5",
		IsDeleted: false,
	}
	_, err := connection.db.Execute(ReadUserQuery, userPass)
	require.Error(t, err, "user account that does not exist")

	// Username and account id collisions.
	for key, testCase := range testUserRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			resp, err := connection.db.Execute(ReadUserQuery, testCase)
			require.NoError(t, err)
			actual := resp.(*model_cassandra.User)
			require.Equal(t, testCase.AccountID, actual.AccountID, "expected account id does not match returned")
			require.Equal(t, testCase.Username, actual.Username, "expected username does not match returned")
		})
	}
}

func TestCreateResponseQuery(t *testing.T) {
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
	// Lock connection to Cassandra cluster.
	connection.mu.Lock()
	defer connection.mu.Unlock()
	// Insert new responses.
	insertTestResponses(t)

	// Non-existent Response.
	nonExistentResponse := &model_cassandra.Response{
		Username:     "user-1",
		Score:        0,
		QuizResponse: nil,
		QuizID:       gocql.TimeUUID(),
	}
	_, err := connection.db.Execute(ReadResponseQuery, nonExistentResponse)
	require.Error(t, err, "user response that does not exist")

	// Username and quiz id collisions.
	for key, testCase := range testResponseRecords {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			resp, err := connection.db.Execute(ReadResponseQuery, testCase)
			require.NoError(t, err)
			actual := resp.(*model_cassandra.Response)
			require.Equalf(t, testCase.QuizID, actual.QuizID, "expected quiz id, %s, does not match returned %s", testCase.QuizID.String(), actual.QuizID.String())
			require.Equalf(t, testCase.Username, actual.Username, "expected username, %s, does not match returned %s", testCase.Username, actual.Username)
		})
	}
}

func TestReadResponseStatisticsQuery(t *testing.T) {
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
			2,
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

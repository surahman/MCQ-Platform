package data_store

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
)

var testUserRecords = GetTestUsers()

func insertTestUsers(t *testing.T) {
	_, err := truncateTableQuery(connection.db, "users")
	require.NoErrorf(t, err, "failed to truncate user table before populating")

	for _, user := range testUserRecords {
		_, err := CreateUserQuery(connection.db, user)
		require.NoErrorf(t, err, "failed to create user %v with error %v", user, err)
	}
}

func freshTestData(t *testing.T) {
	insertTestUsers(t)
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

package data_store

import (
	"fmt"

	"github.com/surahman/mcq-platform/pkg/model/cassandra"
)

// GetTestUsers will generate a number of dummy users for testing.
func GetTestUsers() map[string]*model_cassandra.User {
	users := make(map[string]*model_cassandra.User)
	username := "user-%d"
	accountID := "user-account-id-%d"
	password := "user-pwd-%d"
	firstname := "firstname-%d"
	lastname := "lastname-%d"
	email := "user%d@email-address.com"

	for idx := 1; idx < 5; idx++ {
		uname := fmt.Sprintf(username, idx)
		users[uname] = &model_cassandra.User{
			UserAccount: &model_cassandra.UserAccount{
				Username:  fmt.Sprintf(username, idx),
				Password:  fmt.Sprintf(password, idx),
				FirstName: fmt.Sprintf(firstname, idx),
				LastName:  fmt.Sprintf(lastname, idx),
				Email:     fmt.Sprintf(email, idx),
			},
			AccountID: fmt.Sprintf(accountID, idx),
			IsDeleted: false,
		}
	}

	return users
}

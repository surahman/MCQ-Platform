package model_cassandra

import (
	"encoding/base64"

	"golang.org/x/crypto/blake2b"
)

// User represents a users account and is a row in user table.
type User struct {
	*UserAccount
	AccountID string `json:"account_id,omitempty" cql:"account_id"`
	IsDeleted bool   `json:"is_deleted" cql:"is_deleted"`
}

// UserAccount is the core user account information.
type UserAccount struct {
	Username  string `json:"username,omitempty" cql:"username" validate:"required,min=8,alphanum"`
	Password  string `json:"password,omitempty" cql:"password" validate:"required,min=8,max=32"`
	FirstName string `json:"first_name,omitempty" cql:"first_name" validate:"required"`
	LastName  string `json:"last_name,omitempty" cql:"last_name" validate:"required"`
	Email     string `json:"email,omitempty" cql:"email" validate:"required,email"`
}

// Blake2b256 will create a hash from an input string. This hash is used to create the Account ID for a user.
func Blake2b256(data string) string {
	hash := blake2b.Sum256([]byte(data))
	return base64.URLEncoding.EncodeToString(hash[:])
}

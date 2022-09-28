package model_cassandra

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

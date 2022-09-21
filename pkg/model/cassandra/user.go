package model_cassandra

// User represents an account and is a row in user table.
type User struct {
	*UserAccount
	AccountID string `json:"account_id,omitempty" yaml:"account_id,omitempty"`
	IsDeleted bool   `json:"is_deleted" yaml:"is_deleted"`
}

// UserAccount is the core user account information.
type UserAccount struct {
	Username  string `json:"username,omitempty" yaml:"username,omitempty" validate:"required,min=8,alphanum"`
	Password  string `json:"password,omitempty" yaml:"password,omitempty" validate:"required,min=8,max=32"`
	FirstName string `json:"first_name,omitempty" yaml:"first_name,omitempty" validate:"required"`
	LastName  string `json:"last_name,omitempty" yaml:"last_name,omitempty" validate:"required"`
	Email     string `json:"email,omitempty" yaml:"email,omitempty" validate:"required,email"`
}

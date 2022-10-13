package model_rest

import (
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
)

// JWTAuthResponse is the response to a successful login and token refresh. The expires field is used on by the client to
// know when to refresh the token.
type JWTAuthResponse struct {
	Token   string `json:"token" yaml:"token" validate:"required"`     // JWT string sent too and validated by server.
	Expires int64  `json:"expires" yaml:"expires" validate:"required"` // Expiration time as unix time stamp.
}

// Error is a generic error message that is returned to the requester.
type Error struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	Payload any    `json:"payload,omitempty" yaml:"payload,omitempty"`
}

// Success is a generic success message that is returned to the requester.
type Success struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	Payload any    `json:"payload,omitempty" yaml:"payload,omitempty"`
}

// DeleteUserRequest is the request to mark a user account as deleted. The user must supply their login credentials as
// well as a confirmation message.
type DeleteUserRequest struct {
	model_cassandra.UserLoginCredentials
	Confirmation string `json:"confirmation" yaml:"confirmation" validate:"required"`
}

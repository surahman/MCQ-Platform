package model_rest

import "time"

// JWTAuthResponse is the response to a successful login and token refresh. The expires field is used on by the client to
// know when to refresh the token.
type JWTAuthResponse struct {
	Token   string    `json:"token" yaml:"token" validate:"required"`     // JWT string sent too and validated by server.
	Expires time.Time `json:"expires" yaml:"expires" validate:"required"` // Expiration time, only used by end-user.
}

// Error is a generic error that is returned to the requester.
type Error struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	Payload any    `json:"payload,omitempty" yaml:"payload,omitempty"`
}

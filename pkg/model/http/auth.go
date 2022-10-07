package http

import "time"

// JWTAuthResponse is the response to a successful login and token refresh.
type JWTAuthResponse struct {
	Token   string    `json:"token" yaml:"token" validate:"required"`     // JWT string sent too and validated by server.
	Expires time.Time `json:"expires" yaml:"expires" validate:"required"` // Only used by end-user.
}

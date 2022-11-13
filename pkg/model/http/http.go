package model_http

import (
	"github.com/gocql/gocql"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
)

// JWTAuthResponse is the response to a successful login and token refresh. The expires field is used on by the client to
// know when to refresh the token.
type JWTAuthResponse struct {
	Token     string `json:"token" yaml:"token" validate:"required"`         // JWT string sent to and validated by the server.
	Expires   int64  `json:"expires" yaml:"expires" validate:"required"`     // Expiration time as unix time stamp. Strictly used by client to gauge when to refresh the token.
	Threshold int64  `json:"threshold" yaml:"threshold" validate:"required"` // The window in seconds before expiration during which the token can be refreshed.
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

// Metadata contains information on the statistics request.
type Metadata struct {
	QuizID     gocql.UUID `json:"quiz_id"`
	NumRecords int        `json:"num_records"`
}

// StatsResponse is a paginated response to a request for statistics for a specific quiz.
type StatsResponse struct {
	Records  []*model_cassandra.Response `json:"records"`
	Metadata `json:"metadata,omitempty"`
	Links    struct {
		NextPage string `json:"next_page"`
	} `json:"links,omitempty"`
}

// NextPage is information required to access the next page of data from a GraphQL statistics request.
type NextPage struct {
	Cursor   string `json:"next_page"`
	PageSize int    `json:"page_size"`
}

// StatsResponseGraphQL is a paginated GraphQL response to a request for statistics for a specific quiz.
type StatsResponseGraphQL struct {
	Records  []*model_cassandra.Response `json:"records"`
	Metadata `json:"metadata"`
	NextPage `json:"next_page"`
}

package model_cassandra

import "github.com/gocql/gocql"

// Response represents a response to a quiz and is a row in responses table.
type Response struct {
	Username  string       `json:"username,omitempty" cql:"username,omitempty" validator:"required"`
	QuizID    gocql.UUID   `json:"quiz_id,omitempty" cql:"quiz_id,omitempty" validator:"required"`
	Responses QuizResponse `json:"responses,omitempty" cql:"responses,omitempty" validator:"required"`
	Score     float64      `json:"score,omitempty" cql:"score,omitempty" validator:"required"`
}

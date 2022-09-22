package model_cassandra

import "github.com/gocql/gocql"

// Response represents a response to a quiz and is a row in responses table.
type Response struct {
	Username  string       `json:"username,omitempty" yaml:"username,omitempty" validator:"required"`
	QuizID    gocql.UUID   `json:"quiz_id,omitempty" yaml:"quiz_id,omitempty" validator:"required"`
	Responses QuizResponse `json:"responses,omitempty" yaml:"responses,omitempty" validator:"required"`
	Score     float64      `json:"score,omitempty" yaml:"score,omitempty" validator:"required"`
}

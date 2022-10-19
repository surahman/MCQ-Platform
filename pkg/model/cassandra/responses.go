package model_cassandra

import "github.com/gocql/gocql"

// Response represents a response to a quiz and is a row in responses table.
type Response struct {
	Username      string  `json:"username,omitempty" cql:"username" validator:"required"`
	Score         float64 `json:"score,omitempty" cql:"score" validator:"required"`
	*QuizResponse `validator:"required"`
	QuizID        gocql.UUID `json:"quiz_id,omitempty" cql:"quiz_id" validator:"required"`
}

// QuizResponse
// [1] Can have [0-10] questions answered.
// [2] [0-5] options selected for an answer.
// [3] Answer indices must be valid [0-4].
type QuizResponse struct {
	// The answer card to a quiz. The rows indices are the question numbers and the columns indices are the selected option numbers.
	Responses [][]int32 `json:"responses,omitempty" cql:"responses" validate:"required,min=0,max=10,dive,min=0,max=5,dive,min=0,max=4"`
}

// StatsRequest is a request for statistics for a specific quiz.
type StatsRequest struct {
	QuizID     gocql.UUID
	PageCursor string `json:"page_cursor"`
	PageSize   int    `json:"page_size"`
}

// StatsResponse is a paginated response to a request for statistics for a specific quiz.
type StatsResponse struct {
	Records  []*Response `json:"records"`
	Metadata struct {
		QuizID     gocql.UUID `json:"quiz_id"`
		NumRecords int        `json:"num_records"`
	} `json:"metadata,omitempty"`
	Links struct {
		NextPage string `json:"next_page"`
	} `json:"links,omitempty"`
}

package model_cassandra

import "github.com/gocql/gocql"

// Quiz represents a quiz and is a row in the quizzes table.
// [1] Quiz title is required
// [2] Questions array is required and is valid (1-10 questions).
// [3] Validate all Questions.
type Quiz struct {
	*QuizCore              // The title and questions.
	QuizID      gocql.UUID `json:"quiz_id,omitempty" cql:"quiz_id,omitempty"`           // The unique identifier for the quiz.
	Author      string     `json:"author,omitempty" cql:"author,omitempty"`             // The username of the quiz creator.
	IsPublished bool       `json:"is_published,omitempty" cql:"is_published,omitempty"` // Status indicating whether the quiz can be viewed or taken by other users.
	IsDeleted   bool       `json:"is_deleted,omitempty" cql:"is_deleted,omitempty"`     // Status indicating whether the quiz has been deleted.
}

// Question
// [1] Question description is required.
// [2] Options are all defined and are valid (2-5).
// [3] Answer key is required and is valid (1-5 answers in range 0-4).
// [4] Number of answers is less than or equal to number of options.
type Question struct {
	Description string   `json:"description,omitempty" cql:"is_published,omitempty" validate:"required"`                     // The description that contains the text of the question.
	Options     []string `json:"options,omitempty" cql:"options,omitempty" validate:"required,min=2,max=5"`                  // The available options for the question.
	Answers     []int32  `json:"answers,omitempty" cql:"answers,omitempty" validate:"required,min=1,max=5,dive,min=0,max=4"` // The indices of the options that are correct answers in the question.
}

// QuizCore is the actual data used to create as well as what is presented when viewing a quiz.
type QuizCore struct {
	Title     string     `json:"title,omitempty" cql:"title,omitempty" validate:"required"`                           // The title description of the quiz.
	Questions []Question `json:"questions,omitempty" cql:"questions,omitempty" validate:"required,min=1,max=10,dive"` // A list of questions in the quiz.
}

// QuizResponse
// [1] Can have [0-10] questions answered.
// [2] [0-5] options selected for an answer.
// [3] Answer indices must be valid [0-4].
type QuizResponse struct {
	// The answer card to a quiz. The rows indices are the question numbers and the columns indices are the selected option numbers.
	Responses [][]int32 `json:"responses,omitempty" cql:"responses,omitempty" validate:"required,min=0,max=10,dive,min=0,max=5,dive,min=0,max=4"`
}

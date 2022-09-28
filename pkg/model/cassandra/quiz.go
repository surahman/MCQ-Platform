package model_cassandra

import "github.com/gocql/gocql"

// Quiz represents a quiz and is a row in the quizzes table.
// [1] Quiz title is required
// [2] Questions array is required and is valid (1-10 questions).
// [3] Validate all Questions.
type Quiz struct {
	Author      string     `json:"author,omitempty" cql:"author"` // The username of the quiz creator.
	*QuizCore              // The title and questions.
	QuizID      gocql.UUID `json:"quiz_id,omitempty" cql:"quiz_id"`           // The unique identifier for the quiz.
	IsPublished bool       `json:"is_published,omitempty" cql:"is_published"` // Status indicating whether the quiz can be viewed or taken by other users.
	IsDeleted   bool       `json:"is_deleted,omitempty" cql:"is_deleted"`     // Status indicating whether the quiz has been deleted.
}

// Question
// [1] Question description is required.
// [2] Options are all defined and are valid (2-5).
// [3] Answer key is required and is valid (1-5 answers in range 0-4).
// [4] Number of answers is less than or equal to number of options.
// [5] URI of any assets supplied are URL Encoded.
type Question struct {
	Description string   `json:"description,omitempty" cql:"is_published" validate:"required"`                     // The description that contains the text of the question.
	Asset       string   `json:"asset,omitempty" cql:"asset" validate:"url_encoded"`                               // URI of an asset to be displayed with question.
	Options     []string `json:"options,omitempty" cql:"options" validate:"required,min=2,max=5"`                  // The available options for the question.
	Answers     []int32  `json:"answers,omitempty" cql:"answers" validate:"required,min=1,max=5,dive,min=0,max=4"` // The indices of the options that are correct answers in the question.
}

// QuizCore is the actual data used to create as well as what is presented when viewing a quiz.
type QuizCore struct {
	Title       string     `json:"title,omitempty" cql:"title" validate:"required"`                                                          // The title description of the quiz.
	MarkingType string     `json:"type,omitempty" cql:"marking_type" validate:"oneof='None' 'none' 'Negative' 'negative' 'Binary' 'binary'"` // Marking scheme type can be not marked, negative marking, or all or nothing.
	Questions   []Question `json:"questions,omitempty" cql:"questions" validate:"required,min=1,max=10,dive"`                                // A list of questions in the quiz.
}

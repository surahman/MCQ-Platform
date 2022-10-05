package model_cassandra

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/validator"
)

var question1 = Question{Description: "Description of test 1",
	Asset:   "http%3A%2F%2Fwww.url-encoded.web%2Fthis-is-url-encoded%2F",
	Options: []string{"option 1", "option 2", "option 3", "option 4", "option 5"},
	Answers: []int32{0, 1, 2, 3, 4}}
var question2 = Question{Description: "Description of test 2",
	Asset:   "http%3A%2F%2Fwww.url-encoded.web%2Fthis-is-url-encoded%2F",
	Options: []string{"option 1", "option 2", "option 3", "option 4", "option 5"},
	Answers: []int32{0, 2, 4}}
var question3 = Question{Description: "Description of test 3",
	Options: []string{"option 1", "option 2", "option 3", "option 4", "option 5"},
	Answers: []int32{1, 3}}
var questionNotURLEnc = Question{Description: "Description of test 3",
	Asset:   "http://www.url-encoded.web/this-is-url-encoded/<%az",
	Options: []string{"option 1", "option 2", "option 3", "option 4", "option 5"},
	Answers: []int32{1, 3}}
var questionNoDesc = Question{Description: "",
	Options: []string{"option 1", "option 2", "option 3", "option 4", "option 5"},
	Answers: []int32{0, 1, 2, 3, 4}}
var questionNoOpt = Question{Description: "Question with no options",
	Options: []string{},
	Answers: []int32{0, 1, 2, 3, 4}}
var questionTooManyOpt = Question{Description: "Question with too many options",
	Options: []string{"option 1", "option 2", "option 3", "option 4", "option 5", "option 6"},
	Answers: []int32{0, 1, 2, 3, 4}}
var questionNoAns = Question{Description: "Question without answers",
	Options: []string{"option 1", "option 2", "option 3", "option 4", "option 5"},
	Answers: []int32{}}
var questionTooManyAns = Question{Description: "Question with too many answers",
	Options: []string{"option 1", "option 2", "option 3", "option 4", "option 5"},
	Answers: []int32{0, 1, 2, 3, 4, 4}}
var questionBadAns = Question{Description: "Question with bad index answers",
	Options: []string{"option 1", "option 2", "option 3", "option 4", "option 5"},
	Answers: []int32{0, 1, 2, 3, 5}}
var questionAnsGTOpt = Question{Description: "Question with more answers than options",
	Options: []string{"option 1", "option 2", "option 3", "option 4"},
	Answers: []int32{0, 1, 2, 3, 4}}

var quizValid = QuizCore{Title: "Valid quiz", MarkingType: "Negative", Questions: []*Question{&question1, &question2, &question3}}
var quizInvalidMarking = QuizCore{Title: "Valid quiz", MarkingType: "Invalid", Questions: []*Question{&question1, &question2, &question3}}
var quizNoTitle = QuizCore{Title: "", MarkingType: "Negative", Questions: []*Question{&question1, &question2, &question3}}
var quizEmptyQuestions = QuizCore{Title: "No Questions", MarkingType: "Negative", Questions: []*Question{}}
var quizTooManyQuestions = QuizCore{Title: "Too many questions", MarkingType: "Negative", Questions: []*Question{&question1, &question1,
	&question1, &question1, &question1, &question1, &question1, &question1, &question1, &question1, &question1}}
var quizInvalidQuestions = QuizCore{Title: "Invalid questions", MarkingType: "Negative", Questions: []*Question{&questionNoDesc}}
var quizTooManyAnswers = QuizCore{Title: "Too many answers", MarkingType: "Negative", Questions: []*Question{&questionTooManyAns}}
var quizTooManyOpts = QuizCore{Title: "More answers than options", MarkingType: "Negative", Questions: []*Question{&questionAnsGTOpt}}

func TestValidateQuestion(t *testing.T) {
	testCases := []struct {
		name          string
		questionInput *Question
		expectErr     require.ErrorAssertionFunc
		expectedLen   int
	}{
		{"Valid question 1", &question1, require.NoError, 0},
		{"Valid question 2", &question2, require.NoError, 0},
		{"Valid question 3", &question3, require.NoError, 0},
		{"Not URL encoded", &questionNotURLEnc, require.Error, 1},
		{"No description", &questionNoDesc, require.Error, 1},
		{"No options", &questionNoOpt, require.Error, 1},
		{"Too many options", &questionTooManyOpt, require.Error, 1},
		{"No answers in key", &questionNoAns, require.Error, 1},
		{"Too many answers in key", &questionTooManyAns, require.Error, 1},
		{"Out of range answers in key", &questionBadAns, require.Error, 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validator.ValidateStruct(testCase.questionInput)
			testCase.expectErr(t, err)

			if err != nil {
				require.Equal(t, testCase.expectedLen, len(err.(*validator.ErrorValidation).Errors))
			}
		})
	}
}

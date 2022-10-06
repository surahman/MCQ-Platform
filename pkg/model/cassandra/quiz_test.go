package model_cassandra

import (
	"fmt"
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
var quizValidBinaryMarking = QuizCore{Title: "Valid quiz", MarkingType: "Binary", Questions: []*Question{&question1, &question2, &question3}}
var quizValidNoMarking = QuizCore{Title: "Valid quiz", MarkingType: "None", Questions: []*Question{&question1, &question2, &question3}}
var quizInvalidMarking = QuizCore{Title: "Valid quiz", MarkingType: "Invalid", Questions: []*Question{&question1, &question2, &question3}}
var quizNoTitle = QuizCore{Title: "", MarkingType: "Negative", Questions: []*Question{&question1, &question2, &question3}}
var quizEmptyQuestions = QuizCore{Title: "No Questions", MarkingType: "Negative", Questions: []*Question{}}
var quizTooManyQuestions = QuizCore{Title: "Too many questions", MarkingType: "Negative", Questions: []*Question{&question1, &question1,
	&question1, &question1, &question1, &question1, &question1, &question1, &question1, &question1, &question1}}
var quizInvalidQuestions = QuizCore{Title: "Invalid questions", MarkingType: "Negative", Questions: []*Question{&questionNoDesc}}
var quizTooManyAnswers = QuizCore{Title: "Too many answers", MarkingType: "Negative", Questions: []*Question{&questionTooManyAns}}
var quizTooManyOpts = QuizCore{Title: "More answers than options", MarkingType: "Negative", Questions: []*Question{&questionAnsGTOpt}}

func TestValidateQuestionNum(t *testing.T) {
	testCases := []struct {
		name      string
		input     *Question
		expectErr require.ErrorAssertionFunc
	}{
		// ----- test cases start ----- //
		{
			name: "Valid - equal",
			input: &Question{
				Description: "Valid quiz",
				Asset:       "some-asset",
				Options:     []string{"one", "two", "three", "four", "five"},
				Answers:     []int32{0, 1, 2, 3, 4},
			},
			expectErr: require.NoError,
		}, {
			name: "Valid - fewer answers",
			input: &Question{
				Description: "Valid quiz",
				Asset:       "some-asset",
				Options:     []string{"one", "two", "three", "four", "five"},
				Answers:     []int32{0, 1, 2, 3},
			},
			expectErr: require.NoError,
		}, {
			name: "More answers than options",
			input: &Question{
				Description: "Valid quiz",
				Asset:       "some-asset",
				Options:     []string{"one", "two", "three", "four"},
				Answers:     []int32{0, 1, 2, 3, 4},
			},
			expectErr: require.Error,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validator.ValidateStruct(testCase.input)
			testCase.expectErr(t, err)
		})
	}
}

func TestValidateQuestion(t *testing.T) {
	testCases := []struct {
		name          string
		questionInput *Question
		expectErr     require.ErrorAssertionFunc
		expectedLen   int
	}{
		// ----- test cases start ----- //
		{
			name:          "Valid question 1",
			questionInput: &question1,
			expectErr:     require.NoError,
			expectedLen:   0,
		}, {
			name:          "Valid question 2",
			questionInput: &question2,
			expectErr:     require.NoError,
			expectedLen:   0,
		}, {
			name:          "Valid question 3",
			questionInput: &question3,
			expectErr:     require.NoError,
			expectedLen:   0,
		}, {
			name:          "Not URL encoded",
			questionInput: &questionNotURLEnc,
			expectErr:     require.Error,
			expectedLen:   1,
		}, {
			name:          "No description",
			questionInput: &questionNoDesc,
			expectErr:     require.Error,
			expectedLen:   1,
		}, {
			name:          "No options",
			questionInput: &questionNoOpt,
			expectErr:     require.Error,
			expectedLen:   2,
		}, {
			name:          "Too many options",
			questionInput: &questionTooManyOpt,
			expectErr:     require.Error,
			expectedLen:   1,
		}, {
			name:          "No answers in key",
			questionInput: &questionNoAns,
			expectErr:     require.Error,
			expectedLen:   1,
		}, {
			name:          "Too many answers in key",
			questionInput: &questionTooManyAns,
			expectErr:     require.Error,
			expectedLen:   1,
		}, {
			name:          "Out of range answers in key",
			questionInput: &questionBadAns,
			expectErr:     require.Error,
			expectedLen:   1,
		},
		// ----- test cases end ----- //
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

func TestValidateQuizCore(t *testing.T) {
	testCases := []struct {
		name        string
		input       *QuizCore
		expectErr   require.ErrorAssertionFunc
		expectedLen int
	}{
		// ----- test cases start ----- //
		{
			name:        "Valid question - Negative marking",
			input:       &quizValid,
			expectErr:   require.NoError,
			expectedLen: 0,
		}, {
			name:        "Valid question - Binary marking",
			input:       &quizValidBinaryMarking,
			expectErr:   require.NoError,
			expectedLen: 0,
		}, {
			name:        "Valid question - No marking",
			input:       &quizValidNoMarking,
			expectErr:   require.NoError,
			expectedLen: 0,
		}, {
			name:        "Invalid marking",
			input:       &quizInvalidMarking,
			expectErr:   require.Error,
			expectedLen: 0,
		}, {
			name:        "No title",
			input:       &quizNoTitle,
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name:        "No questions",
			input:       &quizEmptyQuestions,
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name:        "Too many questions",
			input:       &quizTooManyQuestions,
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name:        "Invalid questions",
			input:       &quizInvalidQuestions,
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name:        "Too many answers",
			input:       &quizTooManyAnswers,
			expectErr:   require.Error,
			expectedLen: 1,
		}, {
			name:        "Too many options",
			input:       &quizTooManyOpts,
			expectErr:   require.Error,
			expectedLen: 1,
		},
		// ----- test cases end ----- //
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validator.ValidateStruct(testCase.input)
			testCase.expectErr(t, err)

			if err != nil {
				fmt.Println(len(err.(*validator.ErrorValidation).Errors))
			}
		})
	}
}

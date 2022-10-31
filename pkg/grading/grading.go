package grading

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/surahman/mcq-platform/pkg/model/cassandra"
)

// Mock Grading interface stub generation.
//go:generate mockgen -destination=../mocks/mock_grading.go -package=mocks github.com/surahman/mcq-platform/pkg/grading Grading

// Grading is the interface through which the test grading is facilitated. Created to support mock testing.
type Grading interface {
	// Grade will take grade or mark a quiz response using the answer key provided by the actual quiz.
	Grade(*model_cassandra.QuizResponse, *model_cassandra.QuizCore) (float64, error)
}

// Check to ensure the Cassandra interface has been implemented.
var _ Grading = &gradingImpl{}

// gradingImpl implements the Grading interface and contains the logic for marking functionality.
type gradingImpl struct {
}

// NewGrading creates a new grading instance.
func NewGrading() Grading {
	return newGradingImpl()
}

// newGradingImpl creates a new grading implementation instance.
func newGradingImpl() *gradingImpl {
	return &gradingImpl{}
}

// Grade will mark a quiz response based on the marking type and answer key in the question.
func (g *gradingImpl) Grade(response *model_cassandra.QuizResponse, quiz *model_cassandra.QuizCore) (float64, error) {
	total := 0.0

	// Configure the grading function.
	var gradingFunc func([]int32, map[int32]any, int) float64
	switch strings.ToLower(quiz.MarkingType) {
	case "negative":
		gradingFunc = negativeMarking
	case "non-negative":
		gradingFunc = nonNegativeMarking
	case "binary":
		gradingFunc = binaryMarking
	case "none":
		return math.NaN(), nil
	default:
		return math.NaN(), errors.New("invalid marking type")
	}

	for idx, responses := range response.Responses {
		answerKey := make(map[int32]any)
		for _, val := range quiz.Questions[idx].Answers {
			answerKey[val] = nil
		}

		// Only one answer permitted but multiple provided.
		if len(answerKey) == 1 && len(responses) > 1 {
			errMsg := fmt.Sprintf("only one answer is permitted for: %v", quiz.Questions[idx].Description)
			return math.NaN(), errors.New(errMsg)
		}

		// Grade question.
		total += gradingFunc(responses, answerKey, len(quiz.Questions[idx].Options))
	}

	return total, nil
}

// negativeMarking will grade a question by employing negative marking.
// Marking scheme:
// [Correct] +1 / correct options
// [Wrong] -1 / incorrect options
func negativeMarking(responses []int32, answerKey map[int32]any, numOptions int) float64 {
	totalOptions := numOptions
	correctWeight := float64(len(answerKey))
	incorrectWeight := math.Max(float64(totalOptions)-correctWeight, 1.0) // Division by zero: 1 option for a question results in 0 incorrectResponses.
	correctResponses := 0.0
	incorrectResponses := 0.0

	// Loop over answers and check if they exist.
	for _, val := range responses {
		if _, ok := answerKey[val]; ok {
			correctResponses++
		} else {
			incorrectResponses++
		}
	}

	correctScore := (1.0 / correctWeight) * correctResponses
	incorrectScore := (1.0 / incorrectWeight) * incorrectResponses
	return correctScore - incorrectScore
}

// nonNegativeMarking will grade a question by employing non-negative marking and award partial grades.
// Marking scheme:
// [Correct] +1 / correct options
// [Incorrect] 0 on whole question
func nonNegativeMarking(responses []int32, answerKey map[int32]any, _ int) float64 {
	correctWeight := float64(len(answerKey))
	correctResponses := 0.0
	incorrectResponses := 0.0

	// Loop over answers and check if they exist.
	for _, val := range responses {
		if _, ok := answerKey[val]; ok {
			correctResponses++
		} else {
			incorrectResponses++
		}
	}
	if incorrectResponses == 0 {
		return (1.0 / correctWeight) * correctResponses
	}
	return 0
}

// binaryMarking will validate a result and employ all-or-nothing marking.
// Marking scheme:
// [Correct] +1 if all required answers selected
// [Incorrect] 0 on whole question
func binaryMarking(responses []int32, answerKey map[int32]any, _ int) float64 {
	correctResponses := 0
	incorrectResponses := 0

	// Loop over answers and check if they exist.
	for _, val := range responses {
		if _, ok := answerKey[val]; ok {
			correctResponses++
		} else {
			incorrectResponses++
		}
	}

	if correctResponses == len(answerKey) && incorrectResponses == 0 {
		return 1
	}
	return 0
}

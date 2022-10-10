package grading

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/surahman/mcq-platform/pkg/model/cassandra"
)

func TestNewGrading(t *testing.T) {
	require.NotNil(t, NewGrading(), "failed to create Grading instance")
}

func TestNewGradingImpl(t *testing.T) {
	require.Equalf(t, reflect.TypeOf(gradingImpl{}), reflect.TypeOf(*newGradingImpl()), "")
}

func TestGradingImpl_Marking(t *testing.T) {

	answerKey := map[int32]any{0: nil, 1: nil, 2: nil, 3: nil, 4: nil}

	testCases := []struct {
		name              string
		numOptions        int
		responses         []int32
		expectNegative    float64
		expectNonNegative float64
		expectBinary      float64
	}{
		// ----- test cases start ----- //
		{
			"no responses",
			10,
			[]int32{},
			0,
			0,
			0,
		}, {
			"all correct",
			10,
			[]int32{0, 1, 2, 3, 4},
			1,
			1,
			1,
		}, {
			"all incorrect",
			10,
			[]int32{5, 6, 7, 8, 9},
			-1,
			0,
			0,
		}, {
			"half correct and incorrect",
			10,
			[]int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			0,
			0,
			0,
		}, {
			"partial correct and no incorrect",
			10,
			[]int32{0, 1, 2},
			0.6,
			0.6,
			0,
		}, {
			"partial correct and incorrect",
			10,
			[]int32{0, 1, 2, 8, 9},
			0.2,
			0,
			0,
		}, {
			"no correct and partial incorrect",
			10,
			[]int32{5, 6, 7},
			-0.6,
			0,
			0,
		}, {
			"more incorrect than correct",
			10,
			[]int32{0, 1, 2, 5, 6, 7, 8, 9},
			-0.4,
			0,
			0,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.InDelta(t, testCase.expectBinary, binaryMarking(testCase.responses, answerKey, testCase.numOptions), 0.01, "failed binary marking")
			require.InDelta(t, testCase.expectNegative, negativeMarking(testCase.responses, answerKey, testCase.numOptions), 0.01, "failed negative marking")
			require.InDelta(t, testCase.expectNonNegative, nonNegativeMarking(testCase.responses, answerKey, testCase.numOptions), 0.01, "failed non-negative marking")
		})
	}
}

func TestGradingImpl_Grade(t *testing.T) {
	temperatureQuestion := model_cassandra.Question{Description: "Temperature can be measured in",
		Options: []string{"Kelvin", "Fahrenheit", "Gram", "Celsius", "Liters"},
		Answers: []int32{0, 1, 3}}
	moonQuestion := model_cassandra.Question{Description: "The moon is a star",
		Options: []string{"True", "False"},
		Answers: []int32{1}}
	grader := gradingImpl{}

	testCases := []struct {
		name         string
		expectErrMsg string
		markingType  []string
		questions    []*model_cassandra.Question
		responses    *model_cassandra.QuizResponse
		expectErr    require.ErrorAssertionFunc
		expectScore  []float64
	}{
		// ----- test cases start ----- //
		{
			name:         "invalid marking type",
			expectErrMsg: "invalid",
			markingType:  []string{"invalid"},
			questions:    []*model_cassandra.Question{&temperatureQuestion, &moonQuestion},
			responses:    &model_cassandra.QuizResponse{Responses: [][]int32{{0, 1, 3}, {1}}},
			expectErr:    require.Error,
			expectScore:  []float64{0},
		}, {
			name:         "invalid response - single option question",
			expectErrMsg: "only one answer",
			markingType:  []string{"binary", "negative", "non-negative"},
			questions:    []*model_cassandra.Question{&temperatureQuestion, &moonQuestion},
			responses:    &model_cassandra.QuizResponse{Responses: [][]int32{{0, 1, 3}, {0, 1}}},
			expectErr:    require.Error,
			expectScore:  []float64{0, 0, 0},
		}, {
			name:         "no marking scheme",
			expectErrMsg: "",
			markingType:  []string{"none"},
			questions:    []*model_cassandra.Question{&temperatureQuestion, &moonQuestion},
			responses:    &model_cassandra.QuizResponse{Responses: [][]int32{{0, 1, 3}, {0, 1}}},
			expectErr:    require.NoError,
			expectScore:  []float64{math.NaN()},
		}, {
			name:         "equal correct and incorrect",
			expectErrMsg: "",
			markingType:  []string{"binary", "negative", "non-negative"},
			questions:    []*model_cassandra.Question{&temperatureQuestion, &moonQuestion},
			responses:    &model_cassandra.QuizResponse{Responses: [][]int32{{0, 1, 2, 4}, {1}}},
			expectErr:    require.NoError,
			expectScore:  []float64{1, 0.66, 1},
		}, {
			name:         "all incorrect",
			expectErrMsg: "",
			markingType:  []string{"binary", "negative", "non-negative"},
			questions:    []*model_cassandra.Question{&temperatureQuestion, &moonQuestion},
			responses:    &model_cassandra.QuizResponse{Responses: [][]int32{{2, 4}, {0}}},
			expectErr:    require.NoError,
			expectScore:  []float64{0, -2, 0},
		}, {
			name:         "all correct",
			expectErrMsg: "",
			markingType:  []string{"binary", "negative", "non-negative"},
			questions:    []*model_cassandra.Question{&temperatureQuestion, &moonQuestion},
			responses:    &model_cassandra.QuizResponse{Responses: [][]int32{{0, 1, 3}, {1}}},
			expectErr:    require.NoError,
			expectScore:  []float64{2, 2, 2},
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testQuiz := &model_cassandra.QuizCore{
				Title:       "",
				MarkingType: "",
				Questions:   testCase.questions,
			}

			for idx, markingScheme := range testCase.markingType {
				testQuiz.MarkingType = markingScheme
				score, err := grader.Grade(testCase.responses, testQuiz)
				testCase.expectErr(t, err, "error expectation failed")

				if err != nil {
					require.Contains(t, err.Error(), testCase.expectErrMsg)
					break
				}

				require.InDeltaf(t, testCase.expectScore[idx], score, 0.1, "score delta offset to large for marking scheme %s", markingScheme)
			}
		})
	}
}

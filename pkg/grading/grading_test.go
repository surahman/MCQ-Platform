package grading

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
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

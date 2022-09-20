package utilities

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorField_Error(t *testing.T) {
	const errorStr = "Field: %s, Tag: %s, Value: %s\n"
	testCases := []struct {
		name     string
		input    *ErrorField
		expected string
	}{
		// ----- test cases start ----- //
		{
			"Empty error",
			&ErrorField{},
			fmt.Sprintf(errorStr, "", "", ""),
		},
		{
			"Field only error",
			&ErrorField{Field: "field"},
			fmt.Sprintf(errorStr, "field", "", ""),
		},
		{
			"Field and Tag only error",
			&ErrorField{Field: "field", Tag: "tag"},
			fmt.Sprintf(errorStr, "field", "tag", ""),
		},
		{
			"Field, Tag, and Value error",
			&ErrorField{Field: "field", Tag: "tag", Value: "value"},
			fmt.Sprintf(errorStr, "field", "tag", "value"),
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, testCase.input.Error())
		})
	}
}

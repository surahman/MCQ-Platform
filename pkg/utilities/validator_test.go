package utilities

import (
	"bytes"
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

func TestErrorValidation_Error(t *testing.T) {

	genExpected := func(errs ...string) string {
		var buffer bytes.Buffer
		for _, item := range errs {
			buffer.WriteString(item)
		}
		return buffer.String()
	}

	const errorStr = "Field: %s, Tag: %s, Value: %s\n"

	testCases := []struct {
		name     string
		input    *ErrorValidation
		expected string
	}{
		// ----- test cases start ----- //
		{
			"Empty error",
			&ErrorValidation{Errors: []*ErrorField{}},
			genExpected(""),
		},
		{
			"Single error",
			&ErrorValidation{Errors: []*ErrorField{
				{Field: "field 1", Tag: "tag 1", Value: "value 1"},
			}},
			genExpected(fmt.Sprintf(errorStr, "field 1", "tag 1", "value 1")),
		},
		{
			"Two errors",
			&ErrorValidation{Errors: []*ErrorField{
				{Field: "field 1", Tag: "tag 1", Value: "value 1"},
				{Field: "field 2", Tag: "tag 2", Value: "value 2"},
			}},
			genExpected(fmt.Sprintf(errorStr, "field 1", "tag 1", "value 1"),
				fmt.Sprintf(errorStr, "field 2", "tag 2", "value 2")),
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, testCase.input.Error())
		})
	}
}

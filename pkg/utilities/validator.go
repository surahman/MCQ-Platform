package utilities

import "fmt"

// ErrorField contains information on JSON validation errors.
type ErrorField struct {
	Field string `json:"field" yaml:"field"` // Field name where the validation error occurred.
	Tag   string `json:"tag" yaml:"tag"`     // The reason for the validation failure.
	Value string `json:"value" yaml:"value"` // The value(s) associated with the failure.
}

// Error will output the validation error for a single structs data member.
func (err *ErrorField) Error() string {
	return fmt.Sprintf("Field: %s, Tag: %s, Value: %s\n", err.Field, err.Tag, err.Value)
}

package validator

import (
	"bytes"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
)

// structValidator is the validator instance that is used for structure validation.
var structValidator *validator.Validate

// init the validator and add custom validation rules.
func init() {
	structValidator = validator.New()
	if err := structValidator.RegisterValidation("answers_LT_options", validateNumAnswers); err != nil {
		log.Fatalf("failed to initialize struct validator with custom validation rule: %v", err)
	}
}

// ErrorField contains information on JSON validation errors.
type ErrorField struct {
	Field string `json:"field" yaml:"field"` // Field name where the validation error occurred.
	Tag   string `json:"tag" yaml:"tag"`     // The reason for the validation failure.
	Value any    `json:"value" yaml:"value"` // The value(s) associated with the failure.
}

// Error will output the validation error for a single structs data member.
func (err *ErrorField) Error() string {
	return fmt.Sprintf("Field: %s, Tag: %s, Value: %s\n", err.Field, err.Tag, err.Value)
}

// ErrorValidation contains all the validation errors found in a struct.
type ErrorValidation struct {
	Errors []*ErrorField `json:"validation_errors" yaml:"validation_errors"` // A list of all data members that failed validation.
}

// Error will output the validation error for all struct data members.
func (err *ErrorValidation) Error() string {
	var buffer bytes.Buffer
	for _, item := range err.Errors {
		buffer.WriteString(item.Error())
	}
	return buffer.String()
}

// ValidateStruct will validate a struct and list all deficiencies.
func ValidateStruct(body any) error {
	var validationErr ErrorValidation
	if err := structValidator.Struct(body); err != nil {
		for _, issue := range err.(validator.ValidationErrors) {
			var ev ErrorField
			ev.Field = issue.Field()
			ev.Tag = issue.Tag()
			ev.Value = issue.Value()
			validationErr.Errors = append(validationErr.Errors, &ev)
		}
	}
	if validationErr.Errors == nil {
		return nil
	}
	return &validationErr
}

// validateNumAnswers is used by the validator to check if the number of answers is not greater than the number of questions.
func validateNumAnswers(fieldValue validator.FieldLevel) bool {
	options := fieldValue.Parent().FieldByName("Options").Interface().([]string)
	numOptions := len(options)
	numAnswers := fieldValue.Field().Len()

	if numAnswers > numOptions {
		return false
	}

	return true
}

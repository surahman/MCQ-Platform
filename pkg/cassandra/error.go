package cassandra

import "net/http"

// Error is the internal Cassandra error that uses the builder pattern to assign error codes.
type Error struct {
	Message string
	Status  int
}

// Check to ensure the error interface is implemented.
var _ error = &Error{}

// NewError will generate a new Cassandra error.
func NewError(message string) *Error {
	return &Error{Message: message}
}

// Error returns the message.
func (err *Error) Error() string {
	return err.Message
}

// OK sets the HTTP OK status code.
func (err *Error) OK() *Error {
	err.Status = http.StatusOK
	return err
}

// internalError sets the HTTP Internal Server status code.
func (err *Error) internalError() *Error {
	err.Status = http.StatusInternalServerError
	return err
}

// notFoundError sets the HTTP Not Found status code.
func (err *Error) notFoundError() *Error {
	err.Status = http.StatusNotFound
	return err
}

// conflictError sets the HTTP Conflict status code.
func (err *Error) conflictError() *Error {
	err.Status = http.StatusConflict
	return err
}

// forbiddenError sets the HTTP Forbidden status code.
func (err *Error) forbiddenError() *Error {
	err.Status = http.StatusForbidden
	return err
}

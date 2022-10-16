package cassandra

import (
	"net/http"
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	testCases := []struct {
		name   string
		err    *Error
		status int
	}{
		// ----- test cases start ----- //
		{
			name:   "OK",
			err:    NewError(xid.New().String()).OK(),
			status: http.StatusOK,
		}, {
			name:   "Internal",
			err:    NewError(xid.New().String()).internalError(),
			status: http.StatusInternalServerError,
		}, {
			name:   "Not Found",
			err:    NewError(xid.New().String()).notFoundError(),
			status: http.StatusNotFound,
		}, {
			name:   "Conflict",
			err:    NewError(xid.New().String()).conflictError(),
			status: http.StatusConflict,
		}, {
			name:   "Forbidden",
			err:    NewError(xid.New().String()).forbiddenError(),
			status: http.StatusForbidden,
		},
		// ----- test cases end ----- //
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.status, testCase.status)
		})
	}
}

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

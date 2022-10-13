// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/surahman/mcq-platform/pkg/grading (interfaces: Grading)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model_cassandra "github.com/surahman/mcq-platform/pkg/model/cassandra"
)

// MockGrading is a mock of Grading interface.
type MockGrading struct {
	ctrl     *gomock.Controller
	recorder *MockGradingMockRecorder
}

// MockGradingMockRecorder is the mock recorder for MockGrading.
type MockGradingMockRecorder struct {
	mock *MockGrading
}

// NewMockGrading creates a new mock instance.
func NewMockGrading(ctrl *gomock.Controller) *MockGrading {
	mock := &MockGrading{ctrl: ctrl}
	mock.recorder = &MockGradingMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGrading) EXPECT() *MockGradingMockRecorder {
	return m.recorder
}

// Grade mocks base method.
func (m *MockGrading) Grade(arg0 *model_cassandra.QuizResponse, arg1 *model_cassandra.QuizCore) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Grade", arg0, arg1)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Grade indicates an expected call of Grade.
func (mr *MockGradingMockRecorder) Grade(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Grade", reflect.TypeOf((*MockGrading)(nil).Grade), arg0, arg1)
}
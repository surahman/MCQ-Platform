// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/surahman/mcq-platform/pkg/auth (interfaces: Auth)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model_rest "github.com/surahman/mcq-platform/pkg/model/http"
)

// MockAuth is a mock of Auth interface.
type MockAuth struct {
	ctrl     *gomock.Controller
	recorder *MockAuthMockRecorder
}

// MockAuthMockRecorder is the mock recorder for MockAuth.
type MockAuthMockRecorder struct {
	mock *MockAuth
}

// NewMockAuth creates a new mock instance.
func NewMockAuth(ctrl *gomock.Controller) *MockAuth {
	mock := &MockAuth{ctrl: ctrl}
	mock.recorder = &MockAuthMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuth) EXPECT() *MockAuthMockRecorder {
	return m.recorder
}

// CheckPassword mocks base method.
func (m *MockAuth) CheckPassword(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckPassword", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckPassword indicates an expected call of CheckPassword.
func (mr *MockAuthMockRecorder) CheckPassword(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckPassword", reflect.TypeOf((*MockAuth)(nil).CheckPassword), arg0, arg1)
}

// DecryptFromBytes mocks base method.
func (m *MockAuth) DecryptFromBytes(arg0 []byte) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DecryptFromBytes", arg0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DecryptFromBytes indicates an expected call of DecryptFromBytes.
func (mr *MockAuthMockRecorder) DecryptFromBytes(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DecryptFromBytes", reflect.TypeOf((*MockAuth)(nil).DecryptFromBytes), arg0)
}

// DecryptFromString mocks base method.
func (m *MockAuth) DecryptFromString(arg0 string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DecryptFromString", arg0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DecryptFromString indicates an expected call of DecryptFromString.
func (mr *MockAuthMockRecorder) DecryptFromString(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DecryptFromString", reflect.TypeOf((*MockAuth)(nil).DecryptFromString), arg0)
}

// EncryptToBytes mocks base method.
func (m *MockAuth) EncryptToBytes(arg0 []byte) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EncryptToBytes", arg0)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EncryptToBytes indicates an expected call of EncryptToBytes.
func (mr *MockAuthMockRecorder) EncryptToBytes(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EncryptToBytes", reflect.TypeOf((*MockAuth)(nil).EncryptToBytes), arg0)
}

// EncryptToString mocks base method.
func (m *MockAuth) EncryptToString(arg0 []byte) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EncryptToString", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EncryptToString indicates an expected call of EncryptToString.
func (mr *MockAuthMockRecorder) EncryptToString(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EncryptToString", reflect.TypeOf((*MockAuth)(nil).EncryptToString), arg0)
}

// GenerateJWT mocks base method.
func (m *MockAuth) GenerateJWT(arg0 string) (*model_rest.JWTAuthResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateJWT", arg0)
	ret0, _ := ret[0].(*model_rest.JWTAuthResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateJWT indicates an expected call of GenerateJWT.
func (mr *MockAuthMockRecorder) GenerateJWT(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateJWT", reflect.TypeOf((*MockAuth)(nil).GenerateJWT), arg0)
}

// HashPassword mocks base method.
func (m *MockAuth) HashPassword(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HashPassword", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HashPassword indicates an expected call of HashPassword.
func (mr *MockAuthMockRecorder) HashPassword(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HashPassword", reflect.TypeOf((*MockAuth)(nil).HashPassword), arg0)
}

// RefreshJWT mocks base method.
func (m *MockAuth) RefreshJWT(arg0 string) (*model_rest.JWTAuthResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshJWT", arg0)
	ret0, _ := ret[0].(*model_rest.JWTAuthResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RefreshJWT indicates an expected call of RefreshJWT.
func (mr *MockAuthMockRecorder) RefreshJWT(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshJWT", reflect.TypeOf((*MockAuth)(nil).RefreshJWT), arg0)
}

// RefreshThreshold mocks base method.
func (m *MockAuth) RefreshThreshold() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshThreshold")
	ret0, _ := ret[0].(int64)
	return ret0
}

// RefreshThreshold indicates an expected call of RefreshThreshold.
func (mr *MockAuthMockRecorder) RefreshThreshold() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshThreshold", reflect.TypeOf((*MockAuth)(nil).RefreshThreshold))
}

// ValidateJWT mocks base method.
func (m *MockAuth) ValidateJWT(arg0 string) (string, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateJWT", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ValidateJWT indicates an expected call of ValidateJWT.
func (mr *MockAuthMockRecorder) ValidateJWT(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateJWT", reflect.TypeOf((*MockAuth)(nil).ValidateJWT), arg0)
}

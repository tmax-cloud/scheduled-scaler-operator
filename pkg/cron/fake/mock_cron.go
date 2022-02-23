// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/cron/cron.go

// Package fake is a generated GoMock package.
package fake

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	scaler "github.com/tmax-cloud/scheduled-scaler-operator/pkg/scaler"
)

// MockCron is a mock of Cron interface.
type MockCron struct {
	ctrl     *gomock.Controller
	recorder *MockCronMockRecorder
}

// MockCronMockRecorder is the mock recorder for MockCron.
type MockCronMockRecorder struct {
	mock *MockCron
}

// NewMockCron creates a new mock instance.
func NewMockCron(ctrl *gomock.Controller) *MockCron {
	mock := &MockCron{ctrl: ctrl}
	mock.recorder = &MockCronMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCron) EXPECT() *MockCronMockRecorder {
	return m.recorder
}

// Push mocks base method.
func (m *MockCron) Push(arg0 scaler.Scaler) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Push", arg0)
}

// Push indicates an expected call of Push.
func (mr *MockCronMockRecorder) Push(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Push", reflect.TypeOf((*MockCron)(nil).Push), arg0)
}

// Start mocks base method.
func (m *MockCron) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockCronMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockCron)(nil).Start))
}

// Stop mocks base method.
func (m *MockCron) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockCronMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockCron)(nil).Stop))
}

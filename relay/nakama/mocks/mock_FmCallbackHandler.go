// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	runtime "github.com/heroiclabs/nakama-common/runtime"
	mock "github.com/stretchr/testify/mock"
)

// MockFmCallbackHandler is an autogenerated mock type for the FmCallbackHandler type
type MockFmCallbackHandler struct {
	mock.Mock
}

// GenerateCallbackId provides a mock function with given fields:
func (_m *MockFmCallbackHandler) GenerateCallbackId() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// InvokeCallback provides a mock function with given fields: callbackId, status, instanceInfo, sessionInfo, metadata, err
func (_m *MockFmCallbackHandler) InvokeCallback(callbackId string, status runtime.FmCreateStatus, instanceInfo *runtime.InstanceInfo, sessionInfo []*runtime.SessionInfo, metadata map[string]interface{}, err error) {
	_m.Called(callbackId, status, instanceInfo, sessionInfo, metadata, err)
}

// SetCallback provides a mock function with given fields: callbackId, fn
func (_m *MockFmCallbackHandler) SetCallback(callbackId string, fn runtime.FmCreateCallbackFn) {
	_m.Called(callbackId, fn)
}

type mockConstructorTestingTNewMockFmCallbackHandler interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockFmCallbackHandler creates a new instance of MockFmCallbackHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockFmCallbackHandler(t mockConstructorTestingTNewMockFmCallbackHandler) *MockFmCallbackHandler {
	mock := &MockFmCallbackHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

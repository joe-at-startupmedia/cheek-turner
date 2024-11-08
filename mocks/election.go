// Code generated by mockery v2.46.3. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Election is an autogenerated mock type for the ElectionInterface type
type Election struct {
	mock.Mock
}

// IsLeader provides a mock function with given fields:
func (_m *Election) IsLeader() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsLeader")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Stop provides a mock function with given fields:
func (_m *Election) Stop() {
	_m.Called()
}

// NewElection creates a new instance of Election. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewElection(t interface {
	mock.TestingT
	Cleanup(func())
}) *Election {
	mock := &Election{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

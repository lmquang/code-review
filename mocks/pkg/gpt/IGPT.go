// Code generated by mockery v2.46.0. DO NOT EDIT.

package mocks

import (
	openai "github.com/lmquang/code-review/pkg/gpt/openai"
	mock "github.com/stretchr/testify/mock"
)

// IGPT is an autogenerated mock type for the IGPT type
type IGPT struct {
	mock.Mock
}

// Client provides a mock function with given fields:
func (_m *IGPT) Client() openai.IOpenAI {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Client")
	}

	var r0 openai.IOpenAI
	if rf, ok := ret.Get(0).(func() openai.IOpenAI); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(openai.IOpenAI)
		}
	}

	return r0
}

// Review provides a mock function with given fields: formattedDiff
func (_m *IGPT) Review(formattedDiff string) (string, error) {
	ret := _m.Called(formattedDiff)

	if len(ret) == 0 {
		panic("no return value specified for Review")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(formattedDiff)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(formattedDiff)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(formattedDiff)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewIGPT creates a new instance of IGPT. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIGPT(t interface {
	mock.TestingT
	Cleanup(func())
}) *IGPT {
	mock := &IGPT{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

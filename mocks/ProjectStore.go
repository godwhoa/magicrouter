// Code generated by mockery v2.36.0. DO NOT EDIT.

package mocks

import (
	core "magicrouter/core"

	mock "github.com/stretchr/testify/mock"
)

// ProjectStore is an autogenerated mock type for the ProjectStore type
type ProjectStore struct {
	mock.Mock
}

// GetConfig provides a mock function with given fields: projectID
func (_m *ProjectStore) GetConfig(projectID string) (*core.ProjectConfig, error) {
	ret := _m.Called(projectID)

	var r0 *core.ProjectConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*core.ProjectConfig, error)); ok {
		return rf(projectID)
	}
	if rf, ok := ret.Get(0).(func(string) *core.ProjectConfig); ok {
		r0 = rf(projectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.ProjectConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(projectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewProjectStore creates a new instance of ProjectStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProjectStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *ProjectStore {
	mock := &ProjectStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// Code generated by mockery. DO NOT EDIT.

package vault

import mock "github.com/stretchr/testify/mock"

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

// GetSecrets provides a mock function with given fields: path
func (_m *MockClient) GetSecrets(path string) (Secrets, error) {
	ret := _m.Called(path)

	if len(ret) == 0 {
		panic("no return value specified for GetSecrets")
	}

	var r0 Secrets
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (Secrets, error)); ok {
		return rf(path)
	}
	if rf, ok := ret.Get(0).(func(string) Secrets); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Secrets)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockClient creates a new instance of MockClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockClient {
	mock := &MockClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	context "context"

	jetflow "github.com/mathieupost/jetflow"
	mock "github.com/stretchr/testify/mock"
)

// Storage is an autogenerated mock type for the Storage type
type Storage struct {
	mock.Mock
}

type Storage_Expecter struct {
	mock *mock.Mock
}

func (_m *Storage) EXPECT() *Storage_Expecter {
	return &Storage_Expecter{mock: &_m.Mock}
}

// Commit provides a mock function with given fields: ctx, call
func (_m *Storage) Commit(ctx context.Context, call jetflow.Request) error {
	ret := _m.Called(ctx, call)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, jetflow.Request) error); ok {
		r0 = rf(ctx, call)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_Commit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Commit'
type Storage_Commit_Call struct {
	*mock.Call
}

// Commit is a helper method to define mock.On call
//   - ctx context.Context
//   - call jetflow.Request
func (_e *Storage_Expecter) Commit(ctx interface{}, call interface{}) *Storage_Commit_Call {
	return &Storage_Commit_Call{Call: _e.mock.On("Commit", ctx, call)}
}

func (_c *Storage_Commit_Call) Run(run func(ctx context.Context, call jetflow.Request)) *Storage_Commit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(jetflow.Request))
	})
	return _c
}

func (_c *Storage_Commit_Call) Return(_a0 error) *Storage_Commit_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_Commit_Call) RunAndReturn(run func(context.Context, jetflow.Request) error) *Storage_Commit_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, call
func (_m *Storage) Get(ctx context.Context, call jetflow.Request) (jetflow.OperatorHandler, error) {
	ret := _m.Called(ctx, call)

	var r0 jetflow.OperatorHandler
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, jetflow.Request) (jetflow.OperatorHandler, error)); ok {
		return rf(ctx, call)
	}
	if rf, ok := ret.Get(0).(func(context.Context, jetflow.Request) jetflow.OperatorHandler); ok {
		r0 = rf(ctx, call)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jetflow.OperatorHandler)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, jetflow.Request) error); ok {
		r1 = rf(ctx, call)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Storage_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type Storage_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - call jetflow.Request
func (_e *Storage_Expecter) Get(ctx interface{}, call interface{}) *Storage_Get_Call {
	return &Storage_Get_Call{Call: _e.mock.On("Get", ctx, call)}
}

func (_c *Storage_Get_Call) Run(run func(ctx context.Context, call jetflow.Request)) *Storage_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(jetflow.Request))
	})
	return _c
}

func (_c *Storage_Get_Call) Return(_a0 jetflow.OperatorHandler, _a1 error) *Storage_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Storage_Get_Call) RunAndReturn(run func(context.Context, jetflow.Request) (jetflow.OperatorHandler, error)) *Storage_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Prepare provides a mock function with given fields: ctx, call
func (_m *Storage) Prepare(ctx context.Context, call jetflow.Request) error {
	ret := _m.Called(ctx, call)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, jetflow.Request) error); ok {
		r0 = rf(ctx, call)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_Prepare_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Prepare'
type Storage_Prepare_Call struct {
	*mock.Call
}

// Prepare is a helper method to define mock.On call
//   - ctx context.Context
//   - call jetflow.Request
func (_e *Storage_Expecter) Prepare(ctx interface{}, call interface{}) *Storage_Prepare_Call {
	return &Storage_Prepare_Call{Call: _e.mock.On("Prepare", ctx, call)}
}

func (_c *Storage_Prepare_Call) Run(run func(ctx context.Context, call jetflow.Request)) *Storage_Prepare_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(jetflow.Request))
	})
	return _c
}

func (_c *Storage_Prepare_Call) Return(_a0 error) *Storage_Prepare_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_Prepare_Call) RunAndReturn(run func(context.Context, jetflow.Request) error) *Storage_Prepare_Call {
	_c.Call.Return(run)
	return _c
}

// Rollback provides a mock function with given fields: ctx, call
func (_m *Storage) Rollback(ctx context.Context, call jetflow.Request) error {
	ret := _m.Called(ctx, call)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, jetflow.Request) error); ok {
		r0 = rf(ctx, call)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Storage_Rollback_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Rollback'
type Storage_Rollback_Call struct {
	*mock.Call
}

// Rollback is a helper method to define mock.On call
//   - ctx context.Context
//   - call jetflow.Request
func (_e *Storage_Expecter) Rollback(ctx interface{}, call interface{}) *Storage_Rollback_Call {
	return &Storage_Rollback_Call{Call: _e.mock.On("Rollback", ctx, call)}
}

func (_c *Storage_Rollback_Call) Run(run func(ctx context.Context, call jetflow.Request)) *Storage_Rollback_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(jetflow.Request))
	})
	return _c
}

func (_c *Storage_Rollback_Call) Return(_a0 error) *Storage_Rollback_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Storage_Rollback_Call) RunAndReturn(run func(context.Context, jetflow.Request) error) *Storage_Rollback_Call {
	_c.Call.Return(run)
	return _c
}

// NewStorage creates a new instance of Storage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStorage(t interface {
	mock.TestingT
	Cleanup(func())
}) *Storage {
	mock := &Storage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

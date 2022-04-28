// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"sync"

	"github.com/libsv/payd"
)

// Ensure, that PrivateKeyReaderWriterMock does implement payd.PrivateKeyReaderWriter.
// If this is not the case, regenerate this file with moq.
var _ payd.PrivateKeyReaderWriter = &PrivateKeyReaderWriterMock{}

// PrivateKeyReaderWriterMock is a mock implementation of payd.PrivateKeyReaderWriter.
//
// 	func TestSomethingThatUsesPrivateKeyReaderWriter(t *testing.T) {
//
// 		// make and configure a mocked payd.PrivateKeyReaderWriter
// 		mockedPrivateKeyReaderWriter := &PrivateKeyReaderWriterMock{
// 			PrivateKeyFunc: func(ctx context.Context, args payd.KeyArgs) (*payd.PrivateKey, error) {
// 				panic("mock out the PrivateKey method")
// 			},
// 			PrivateKeyCreateFunc: func(ctx context.Context, req payd.PrivateKey) (*payd.PrivateKey, error) {
// 				panic("mock out the PrivateKeyCreate method")
// 			},
// 		}
//
// 		// use mockedPrivateKeyReaderWriter in code that requires payd.PrivateKeyReaderWriter
// 		// and then make assertions.
//
// 	}
type PrivateKeyReaderWriterMock struct {
	// PrivateKeyFunc mocks the PrivateKey method.
	PrivateKeyFunc func(ctx context.Context, args payd.KeyArgs) (*payd.PrivateKey, error)

	// PrivateKeyCreateFunc mocks the PrivateKeyCreate method.
	PrivateKeyCreateFunc func(ctx context.Context, req payd.PrivateKey) (*payd.PrivateKey, error)

	// calls tracks calls to the methods.
	calls struct {
		// PrivateKey holds details about calls to the PrivateKey method.
		PrivateKey []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args payd.KeyArgs
		}
		// PrivateKeyCreate holds details about calls to the PrivateKeyCreate method.
		PrivateKeyCreate []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Req is the req argument value.
			Req payd.PrivateKey
		}
	}
	lockPrivateKey       sync.RWMutex
	lockPrivateKeyCreate sync.RWMutex
}

// PrivateKey calls PrivateKeyFunc.
func (mock *PrivateKeyReaderWriterMock) PrivateKey(ctx context.Context, args payd.KeyArgs) (*payd.PrivateKey, error) {
	if mock.PrivateKeyFunc == nil {
		panic("PrivateKeyReaderWriterMock.PrivateKeyFunc: method is nil but PrivateKeyReaderWriter.PrivateKey was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args payd.KeyArgs
	}{
		Ctx:  ctx,
		Args: args,
	}
	mock.lockPrivateKey.Lock()
	mock.calls.PrivateKey = append(mock.calls.PrivateKey, callInfo)
	mock.lockPrivateKey.Unlock()
	return mock.PrivateKeyFunc(ctx, args)
}

// PrivateKeyCalls gets all the calls that were made to PrivateKey.
// Check the length with:
//     len(mockedPrivateKeyReaderWriter.PrivateKeyCalls())
func (mock *PrivateKeyReaderWriterMock) PrivateKeyCalls() []struct {
	Ctx  context.Context
	Args payd.KeyArgs
} {
	var calls []struct {
		Ctx  context.Context
		Args payd.KeyArgs
	}
	mock.lockPrivateKey.RLock()
	calls = mock.calls.PrivateKey
	mock.lockPrivateKey.RUnlock()
	return calls
}

// PrivateKeyCreate calls PrivateKeyCreateFunc.
func (mock *PrivateKeyReaderWriterMock) PrivateKeyCreate(ctx context.Context, req payd.PrivateKey) (*payd.PrivateKey, error) {
	if mock.PrivateKeyCreateFunc == nil {
		panic("PrivateKeyReaderWriterMock.PrivateKeyCreateFunc: method is nil but PrivateKeyReaderWriter.PrivateKeyCreate was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Req payd.PrivateKey
	}{
		Ctx: ctx,
		Req: req,
	}
	mock.lockPrivateKeyCreate.Lock()
	mock.calls.PrivateKeyCreate = append(mock.calls.PrivateKeyCreate, callInfo)
	mock.lockPrivateKeyCreate.Unlock()
	return mock.PrivateKeyCreateFunc(ctx, req)
}

// PrivateKeyCreateCalls gets all the calls that were made to PrivateKeyCreate.
// Check the length with:
//     len(mockedPrivateKeyReaderWriter.PrivateKeyCreateCalls())
func (mock *PrivateKeyReaderWriterMock) PrivateKeyCreateCalls() []struct {
	Ctx context.Context
	Req payd.PrivateKey
} {
	var calls []struct {
		Ctx context.Context
		Req payd.PrivateKey
	}
	mock.lockPrivateKeyCreate.RLock()
	calls = mock.calls.PrivateKeyCreate
	mock.lockPrivateKeyCreate.RUnlock()
	return calls
}

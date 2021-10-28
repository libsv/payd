// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"sync"

	"github.com/libsv/payd"
)

// Ensure, that TransactionWriterMock does implement payd.TransactionWriter.
// If this is not the case, regenerate this file with moq.
var _ payd.TransactionWriter = &TransactionWriterMock{}

// TransactionWriterMock is a mock implementation of payd.TransactionWriter.
//
// 	func TestSomethingThatUsesTransactionWriter(t *testing.T) {
//
// 		// make and configure a mocked payd.TransactionWriter
// 		mockedTransactionWriter := &TransactionWriterMock{
// 			TransactionCreateFunc: func(ctx context.Context, req payd.TransactionCreate) error {
// 				panic("mock out the TransactionCreate method")
// 			},
// 			TransactionUpdateStateFunc: func(ctx context.Context, args payd.TransactionArgs, req payd.TransactionStateUpdate) error {
// 				panic("mock out the TransactionUpdateState method")
// 			},
// 		}
//
// 		// use mockedTransactionWriter in code that requires payd.TransactionWriter
// 		// and then make assertions.
//
// 	}
type TransactionWriterMock struct {
	// TransactionCreateFunc mocks the TransactionCreate method.
	TransactionCreateFunc func(ctx context.Context, req payd.TransactionCreate) error

	// TransactionUpdateStateFunc mocks the TransactionUpdateState method.
	TransactionUpdateStateFunc func(ctx context.Context, args payd.TransactionArgs, req payd.TransactionStateUpdate) error

	// calls tracks calls to the methods.
	calls struct {
		// TransactionCreate holds details about calls to the TransactionCreate method.
		TransactionCreate []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Req is the req argument value.
			Req payd.TransactionCreate
		}
		// TransactionUpdateState holds details about calls to the TransactionUpdateState method.
		TransactionUpdateState []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args payd.TransactionArgs
			// Req is the req argument value.
			Req payd.TransactionStateUpdate
		}
	}
	lockTransactionCreate      sync.RWMutex
	lockTransactionUpdateState sync.RWMutex
}

// TransactionCreate calls TransactionCreateFunc.
func (mock *TransactionWriterMock) TransactionCreate(ctx context.Context, req payd.TransactionCreate) error {
	if mock.TransactionCreateFunc == nil {
		panic("TransactionWriterMock.TransactionCreateFunc: method is nil but TransactionWriter.TransactionCreate was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Req payd.TransactionCreate
	}{
		Ctx: ctx,
		Req: req,
	}
	mock.lockTransactionCreate.Lock()
	mock.calls.TransactionCreate = append(mock.calls.TransactionCreate, callInfo)
	mock.lockTransactionCreate.Unlock()
	return mock.TransactionCreateFunc(ctx, req)
}

// TransactionCreateCalls gets all the calls that were made to TransactionCreate.
// Check the length with:
//     len(mockedTransactionWriter.TransactionCreateCalls())
func (mock *TransactionWriterMock) TransactionCreateCalls() []struct {
	Ctx context.Context
	Req payd.TransactionCreate
} {
	var calls []struct {
		Ctx context.Context
		Req payd.TransactionCreate
	}
	mock.lockTransactionCreate.RLock()
	calls = mock.calls.TransactionCreate
	mock.lockTransactionCreate.RUnlock()
	return calls
}

// TransactionUpdateState calls TransactionUpdateStateFunc.
func (mock *TransactionWriterMock) TransactionUpdateState(ctx context.Context, args payd.TransactionArgs, req payd.TransactionStateUpdate) error {
	if mock.TransactionUpdateStateFunc == nil {
		panic("TransactionWriterMock.TransactionUpdateStateFunc: method is nil but TransactionWriter.TransactionUpdateState was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args payd.TransactionArgs
		Req  payd.TransactionStateUpdate
	}{
		Ctx:  ctx,
		Args: args,
		Req:  req,
	}
	mock.lockTransactionUpdateState.Lock()
	mock.calls.TransactionUpdateState = append(mock.calls.TransactionUpdateState, callInfo)
	mock.lockTransactionUpdateState.Unlock()
	return mock.TransactionUpdateStateFunc(ctx, args, req)
}

// TransactionUpdateStateCalls gets all the calls that were made to TransactionUpdateState.
// Check the length with:
//     len(mockedTransactionWriter.TransactionUpdateStateCalls())
func (mock *TransactionWriterMock) TransactionUpdateStateCalls() []struct {
	Ctx  context.Context
	Args payd.TransactionArgs
	Req  payd.TransactionStateUpdate
} {
	var calls []struct {
		Ctx  context.Context
		Args payd.TransactionArgs
		Req  payd.TransactionStateUpdate
	}
	mock.lockTransactionUpdateState.RLock()
	calls = mock.calls.TransactionUpdateState
	mock.lockTransactionUpdateState.RUnlock()
	return calls
}

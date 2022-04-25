// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"sync"

	"github.com/libsv/go-dpp"
	"github.com/libsv/payd"
)

// Ensure, that ProofCallbackWriterMock does implement payd.ProofCallbackWriter.
// If this is not the case, regenerate this file with moq.
var _ payd.ProofCallbackWriter = &ProofCallbackWriterMock{}

// ProofCallbackWriterMock is a mock implementation of payd.ProofCallbackWriter.
//
// 	func TestSomethingThatUsesProofCallbackWriter(t *testing.T) {
//
// 		// make and configure a mocked payd.ProofCallbackWriter
// 		mockedProofCallbackWriter := &ProofCallbackWriterMock{
// 			ProofCallBacksCreateFunc: func(ctx context.Context, args payd.ProofCallbackArgs, callbacks map[string]dpp.ProofCallback) error {
// 				panic("mock out the ProofCallBacksCreate method")
// 			},
// 		}
//
// 		// use mockedProofCallbackWriter in code that requires payd.ProofCallbackWriter
// 		// and then make assertions.
//
// 	}
type ProofCallbackWriterMock struct {
	// ProofCallBacksCreateFunc mocks the ProofCallBacksCreate method.
	ProofCallBacksCreateFunc func(ctx context.Context, args payd.ProofCallbackArgs, callbacks map[string]dpp.ProofCallback) error

	// calls tracks calls to the methods.
	calls struct {
		// ProofCallBacksCreate holds details about calls to the ProofCallBacksCreate method.
		ProofCallBacksCreate []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args payd.ProofCallbackArgs
			// Callbacks is the callbacks argument value.
			Callbacks map[string]dpp.ProofCallback
		}
	}
	lockProofCallBacksCreate sync.RWMutex
}

// ProofCallBacksCreate calls ProofCallBacksCreateFunc.
func (mock *ProofCallbackWriterMock) ProofCallBacksCreate(ctx context.Context, args payd.ProofCallbackArgs, callbacks map[string]dpp.ProofCallback) error {
	if mock.ProofCallBacksCreateFunc == nil {
		panic("ProofCallbackWriterMock.ProofCallBacksCreateFunc: method is nil but ProofCallbackWriter.ProofCallBacksCreate was just called")
	}
	callInfo := struct {
		Ctx       context.Context
		Args      payd.ProofCallbackArgs
		Callbacks map[string]dpp.ProofCallback
	}{
		Ctx:       ctx,
		Args:      args,
		Callbacks: callbacks,
	}
	mock.lockProofCallBacksCreate.Lock()
	mock.calls.ProofCallBacksCreate = append(mock.calls.ProofCallBacksCreate, callInfo)
	mock.lockProofCallBacksCreate.Unlock()
	return mock.ProofCallBacksCreateFunc(ctx, args, callbacks)
}

// ProofCallBacksCreateCalls gets all the calls that were made to ProofCallBacksCreate.
// Check the length with:
//     len(mockedProofCallbackWriter.ProofCallBacksCreateCalls())
func (mock *ProofCallbackWriterMock) ProofCallBacksCreateCalls() []struct {
	Ctx       context.Context
	Args      payd.ProofCallbackArgs
	Callbacks map[string]dpp.ProofCallback
} {
	var calls []struct {
		Ctx       context.Context
		Args      payd.ProofCallbackArgs
		Callbacks map[string]dpp.ProofCallback
	}
	mock.lockProofCallBacksCreate.RLock()
	calls = mock.calls.ProofCallBacksCreate
	mock.lockProofCallBacksCreate.RUnlock()
	return calls
}

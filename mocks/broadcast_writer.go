// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"sync"

	"github.com/libsv/go-bt/v2"
	"github.com/libsv/payd"
)

// Ensure, that BroadcastWriterMock does implement payd.BroadcastWriter.
// If this is not the case, regenerate this file with moq.
var _ payd.BroadcastWriter = &BroadcastWriterMock{}

// BroadcastWriterMock is a mock implementation of payd.BroadcastWriter.
//
// 	func TestSomethingThatUsesBroadcastWriter(t *testing.T) {
//
// 		// make and configure a mocked payd.BroadcastWriter
// 		mockedBroadcastWriter := &BroadcastWriterMock{
// 			BroadcastFunc: func(ctx context.Context, args payd.BroadcastArgs, tx *bt.Tx) error {
// 				panic("mock out the Broadcast method")
// 			},
// 		}
//
// 		// use mockedBroadcastWriter in code that requires payd.BroadcastWriter
// 		// and then make assertions.
//
// 	}
type BroadcastWriterMock struct {
	// BroadcastFunc mocks the Broadcast method.
	BroadcastFunc func(ctx context.Context, args payd.BroadcastArgs, tx *bt.Tx) error

	// calls tracks calls to the methods.
	calls struct {
		// Broadcast holds details about calls to the Broadcast method.
		Broadcast []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args payd.BroadcastArgs
			// Tx is the tx argument value.
			Tx *bt.Tx
		}
	}
	lockBroadcast sync.RWMutex
}

// Broadcast calls BroadcastFunc.
func (mock *BroadcastWriterMock) Broadcast(ctx context.Context, args payd.BroadcastArgs, tx *bt.Tx) error {
	if mock.BroadcastFunc == nil {
		panic("BroadcastWriterMock.BroadcastFunc: method is nil but BroadcastWriter.Broadcast was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args payd.BroadcastArgs
		Tx   *bt.Tx
	}{
		Ctx:  ctx,
		Args: args,
		Tx:   tx,
	}
	mock.lockBroadcast.Lock()
	mock.calls.Broadcast = append(mock.calls.Broadcast, callInfo)
	mock.lockBroadcast.Unlock()
	return mock.BroadcastFunc(ctx, args, tx)
}

// BroadcastCalls gets all the calls that were made to Broadcast.
// Check the length with:
//     len(mockedBroadcastWriter.BroadcastCalls())
func (mock *BroadcastWriterMock) BroadcastCalls() []struct {
	Ctx  context.Context
	Args payd.BroadcastArgs
	Tx   *bt.Tx
} {
	var calls []struct {
		Ctx  context.Context
		Args payd.BroadcastArgs
		Tx   *bt.Tx
	}
	mock.lockBroadcast.RLock()
	calls = mock.calls.Broadcast
	mock.lockBroadcast.RUnlock()
	return calls
}

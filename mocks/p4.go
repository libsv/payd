// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/libsv/payd"
	"github.com/libsv/payd/data/http"
	"sync"
)

// Ensure, that P4Mock does implement http.P4.
// If this is not the case, regenerate this file with moq.
var _ http.P4 = &P4Mock{}

// P4Mock is a mock implementation of http.P4.
//
// 	func TestSomethingThatUsesP4(t *testing.T) {
//
// 		// make and configure a mocked http.P4
// 		mockedP4 := &P4Mock{
// 			PaymentRequestFunc: func(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error) {
// 				panic("mock out the PaymentRequest method")
// 			},
// 			PaymentSendFunc: func(ctx context.Context, args payd.PayRequest, req payd.PaymentSend) (*payd.PaymentACK, error) {
// 				panic("mock out the PaymentSend method")
// 			},
// 		}
//
// 		// use mockedP4 in code that requires http.P4
// 		// and then make assertions.
//
// 	}
type P4Mock struct {
	// PaymentRequestFunc mocks the PaymentRequest method.
	PaymentRequestFunc func(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error)

	// PaymentSendFunc mocks the PaymentSend method.
	PaymentSendFunc func(ctx context.Context, args payd.PayRequest, req payd.PaymentSend) (*payd.PaymentACK, error)

	// calls tracks calls to the methods.
	calls struct {
		// PaymentRequest holds details about calls to the PaymentRequest method.
		PaymentRequest []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Req is the req argument value.
			Req payd.PayRequest
		}
		// PaymentSend holds details about calls to the PaymentSend method.
		PaymentSend []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args payd.PayRequest
			// Req is the req argument value.
			Req payd.PaymentSend
		}
	}
	lockPaymentRequest sync.RWMutex
	lockPaymentSend    sync.RWMutex
}

// PaymentRequest calls PaymentRequestFunc.
func (mock *P4Mock) PaymentRequest(ctx context.Context, req payd.PayRequest) (*payd.PaymentRequestResponse, error) {
	if mock.PaymentRequestFunc == nil {
		panic("P4Mock.PaymentRequestFunc: method is nil but P4.PaymentRequest was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Req payd.PayRequest
	}{
		Ctx: ctx,
		Req: req,
	}
	mock.lockPaymentRequest.Lock()
	mock.calls.PaymentRequest = append(mock.calls.PaymentRequest, callInfo)
	mock.lockPaymentRequest.Unlock()
	return mock.PaymentRequestFunc(ctx, req)
}

// PaymentRequestCalls gets all the calls that were made to PaymentRequest.
// Check the length with:
//     len(mockedP4.PaymentRequestCalls())
func (mock *P4Mock) PaymentRequestCalls() []struct {
	Ctx context.Context
	Req payd.PayRequest
} {
	var calls []struct {
		Ctx context.Context
		Req payd.PayRequest
	}
	mock.lockPaymentRequest.RLock()
	calls = mock.calls.PaymentRequest
	mock.lockPaymentRequest.RUnlock()
	return calls
}

// PaymentSend calls PaymentSendFunc.
func (mock *P4Mock) PaymentSend(ctx context.Context, args payd.PayRequest, req payd.PaymentSend) (*payd.PaymentACK, error) {
	if mock.PaymentSendFunc == nil {
		panic("P4Mock.PaymentSendFunc: method is nil but P4.PaymentSend was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args payd.PayRequest
		Req  payd.PaymentSend
	}{
		Ctx:  ctx,
		Args: args,
		Req:  req,
	}
	mock.lockPaymentSend.Lock()
	mock.calls.PaymentSend = append(mock.calls.PaymentSend, callInfo)
	mock.lockPaymentSend.Unlock()
	return mock.PaymentSendFunc(ctx, args, req)
}

// PaymentSendCalls gets all the calls that were made to PaymentSend.
// Check the length with:
//     len(mockedP4.PaymentSendCalls())
func (mock *P4Mock) PaymentSendCalls() []struct {
	Ctx  context.Context
	Args payd.PayRequest
	Req  payd.PaymentSend
} {
	var calls []struct {
		Ctx  context.Context
		Args payd.PayRequest
		Req  payd.PaymentSend
	}
	mock.lockPaymentSend.RLock()
	calls = mock.calls.PaymentSend
	mock.lockPaymentSend.RUnlock()
	return calls
}

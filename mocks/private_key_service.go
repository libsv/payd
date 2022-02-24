// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/payd"
	"sync"
)

// Ensure, that PrivateKeyServiceMock does implement payd.PrivateKeyService.
// If this is not the case, regenerate this file with moq.
var _ payd.PrivateKeyService = &PrivateKeyServiceMock{}

// PrivateKeyServiceMock is a mock implementation of payd.PrivateKeyService.
//
// 	func TestSomethingThatUsesPrivateKeyService(t *testing.T) {
//
// 		// make and configure a mocked payd.PrivateKeyService
// 		mockedPrivateKeyService := &PrivateKeyServiceMock{
// 			CreateFunc: func(ctx context.Context, keyName string, userID uint64) error {
// 				panic("mock out the Create method")
// 			},
// 			PrivateKeyFunc: func(ctx context.Context, keyName string, userID uint64) (*bip32.ExtendedKey, error) {
// 				panic("mock out the PrivateKey method")
// 			},
// 		}
//
// 		// use mockedPrivateKeyService in code that requires payd.PrivateKeyService
// 		// and then make assertions.
//
// 	}
type PrivateKeyServiceMock struct {
	// CreateFunc mocks the Create method.
	CreateFunc func(ctx context.Context, keyName string, userID uint64) error

	// PrivateKeyFunc mocks the PrivateKey method.
	PrivateKeyFunc func(ctx context.Context, keyName string, userID uint64) (*bip32.ExtendedKey, error)

	// calls tracks calls to the methods.
	calls struct {
		// Create holds details about calls to the Create method.
		Create []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// KeyName is the keyName argument value.
			KeyName string
			// UserID is the userID argument value.
			UserID uint64
		}
		// PrivateKey holds details about calls to the PrivateKey method.
		PrivateKey []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// KeyName is the keyName argument value.
			KeyName string
			// UserID is the userID argument value.
			UserID uint64
		}
	}
	lockCreate     sync.RWMutex
	lockPrivateKey sync.RWMutex
}

// Create calls CreateFunc.
func (mock *PrivateKeyServiceMock) Create(ctx context.Context, keyName string, userID uint64) error {
	if mock.CreateFunc == nil {
		panic("PrivateKeyServiceMock.CreateFunc: method is nil but PrivateKeyService.Create was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		KeyName string
		UserID  uint64
	}{
		Ctx:     ctx,
		KeyName: keyName,
		UserID:  userID,
	}
	mock.lockCreate.Lock()
	mock.calls.Create = append(mock.calls.Create, callInfo)
	mock.lockCreate.Unlock()
	return mock.CreateFunc(ctx, keyName, userID)
}

// CreateCalls gets all the calls that were made to Create.
// Check the length with:
//     len(mockedPrivateKeyService.CreateCalls())
func (mock *PrivateKeyServiceMock) CreateCalls() []struct {
	Ctx     context.Context
	KeyName string
	UserID  uint64
} {
	var calls []struct {
		Ctx     context.Context
		KeyName string
		UserID  uint64
	}
	mock.lockCreate.RLock()
	calls = mock.calls.Create
	mock.lockCreate.RUnlock()
	return calls
}

// PrivateKey calls PrivateKeyFunc.
func (mock *PrivateKeyServiceMock) PrivateKey(ctx context.Context, keyName string, userID uint64) (*bip32.ExtendedKey, error) {
	if mock.PrivateKeyFunc == nil {
		panic("PrivateKeyServiceMock.PrivateKeyFunc: method is nil but PrivateKeyService.PrivateKey was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		KeyName string
		UserID  uint64
	}{
		Ctx:     ctx,
		KeyName: keyName,
		UserID:  userID,
	}
	mock.lockPrivateKey.Lock()
	mock.calls.PrivateKey = append(mock.calls.PrivateKey, callInfo)
	mock.lockPrivateKey.Unlock()
	return mock.PrivateKeyFunc(ctx, keyName, userID)
}

// PrivateKeyCalls gets all the calls that were made to PrivateKey.
// Check the length with:
//     len(mockedPrivateKeyService.PrivateKeyCalls())
func (mock *PrivateKeyServiceMock) PrivateKeyCalls() []struct {
	Ctx     context.Context
	KeyName string
	UserID  uint64
} {
	var calls []struct {
		Ctx     context.Context
		KeyName string
		UserID  uint64
	}
	mock.lockPrivateKey.RLock()
	calls = mock.calls.PrivateKey
	mock.lockPrivateKey.RUnlock()
	return calls
}

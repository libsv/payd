// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"sync"

	"github.com/libsv/payd"
)

// Ensure, that PeerChannelsStoreMock does implement payd.PeerChannelsStore.
// If this is not the case, regenerate this file with moq.
var _ payd.PeerChannelsStore = &PeerChannelsStoreMock{}

// PeerChannelsStoreMock is a mock implementation of payd.PeerChannelsStore.
//
// 	func TestSomethingThatUsesPeerChannelsStore(t *testing.T) {
//
// 		// make and configure a mocked payd.PeerChannelsStore
// 		mockedPeerChannelsStore := &PeerChannelsStoreMock{
// 			PeerChannelAPITokenCreateFunc: func(ctx context.Context, args *payd.PeerChannelAPITokenStoreArgs) error {
// 				panic("mock out the PeerChannelAPITokenCreate method")
// 			},
// 			PeerChannelAPITokensCreateFunc: func(ctx context.Context, args ...*payd.PeerChannelAPITokenStoreArgs) error {
// 				panic("mock out the PeerChannelAPITokensCreate method")
// 			},
// 			PeerChannelAccountFunc: func(ctx context.Context, args *payd.PeerChannelIDArgs) (*payd.PeerChannelAccount, error) {
// 				panic("mock out the PeerChannelAccount method")
// 			},
// 			PeerChannelCloseChannelFunc: func(ctx context.Context, channelID string) error {
// 				panic("mock out the PeerChannelCloseChannel method")
// 			},
// 			PeerChannelCreateFunc: func(ctx context.Context, args *payd.PeerChannelCreateArgs) error {
// 				panic("mock out the PeerChannelCreate method")
// 			},
// 			PeerChannelsOpenedFunc: func(ctx context.Context, channelType payd.PeerChannelHandlerType) ([]payd.PeerChannel, error) {
// 				panic("mock out the PeerChannelsOpened method")
// 			},
// 		}
//
// 		// use mockedPeerChannelsStore in code that requires payd.PeerChannelsStore
// 		// and then make assertions.
//
// 	}
type PeerChannelsStoreMock struct {
	// PeerChannelAPITokenCreateFunc mocks the PeerChannelAPITokenCreate method.
	PeerChannelAPITokenCreateFunc func(ctx context.Context, args *payd.PeerChannelAPITokenStoreArgs) error

	// PeerChannelAPITokensCreateFunc mocks the PeerChannelAPITokensCreate method.
	PeerChannelAPITokensCreateFunc func(ctx context.Context, args ...*payd.PeerChannelAPITokenStoreArgs) error

	// PeerChannelAccountFunc mocks the PeerChannelAccount method.
	PeerChannelAccountFunc func(ctx context.Context, args *payd.PeerChannelIDArgs) (*payd.PeerChannelAccount, error)

	// PeerChannelCloseChannelFunc mocks the PeerChannelCloseChannel method.
	PeerChannelCloseChannelFunc func(ctx context.Context, channelID string) error

	// PeerChannelCreateFunc mocks the PeerChannelCreate method.
	PeerChannelCreateFunc func(ctx context.Context, args *payd.PeerChannelCreateArgs) error

	// PeerChannelsOpenedFunc mocks the PeerChannelsOpened method.
	PeerChannelsOpenedFunc func(ctx context.Context, channelType payd.PeerChannelHandlerType) ([]payd.PeerChannel, error)

	// calls tracks calls to the methods.
	calls struct {
		// PeerChannelAPITokenCreate holds details about calls to the PeerChannelAPITokenCreate method.
		PeerChannelAPITokenCreate []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args *payd.PeerChannelAPITokenStoreArgs
		}
		// PeerChannelAPITokensCreate holds details about calls to the PeerChannelAPITokensCreate method.
		PeerChannelAPITokensCreate []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args []*payd.PeerChannelAPITokenStoreArgs
		}
		// PeerChannelAccount holds details about calls to the PeerChannelAccount method.
		PeerChannelAccount []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args *payd.PeerChannelIDArgs
		}
		// PeerChannelCloseChannel holds details about calls to the PeerChannelCloseChannel method.
		PeerChannelCloseChannel []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ChannelID is the channelID argument value.
			ChannelID string
		}
		// PeerChannelCreate holds details about calls to the PeerChannelCreate method.
		PeerChannelCreate []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args *payd.PeerChannelCreateArgs
		}
		// PeerChannelsOpened holds details about calls to the PeerChannelsOpened method.
		PeerChannelsOpened []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ChannelType is the channelType argument value.
			ChannelType payd.PeerChannelHandlerType
		}
	}
	lockPeerChannelAPITokenCreate  sync.RWMutex
	lockPeerChannelAPITokensCreate sync.RWMutex
	lockPeerChannelAccount         sync.RWMutex
	lockPeerChannelCloseChannel    sync.RWMutex
	lockPeerChannelCreate          sync.RWMutex
	lockPeerChannelsOpened         sync.RWMutex
}

// PeerChannelAPITokenCreate calls PeerChannelAPITokenCreateFunc.
func (mock *PeerChannelsStoreMock) PeerChannelAPITokenCreate(ctx context.Context, args *payd.PeerChannelAPITokenStoreArgs) error {
	if mock.PeerChannelAPITokenCreateFunc == nil {
		panic("PeerChannelsStoreMock.PeerChannelAPITokenCreateFunc: method is nil but PeerChannelsStore.PeerChannelAPITokenCreate was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args *payd.PeerChannelAPITokenStoreArgs
	}{
		Ctx:  ctx,
		Args: args,
	}
	mock.lockPeerChannelAPITokenCreate.Lock()
	mock.calls.PeerChannelAPITokenCreate = append(mock.calls.PeerChannelAPITokenCreate, callInfo)
	mock.lockPeerChannelAPITokenCreate.Unlock()
	return mock.PeerChannelAPITokenCreateFunc(ctx, args)
}

// PeerChannelAPITokenCreateCalls gets all the calls that were made to PeerChannelAPITokenCreate.
// Check the length with:
//     len(mockedPeerChannelsStore.PeerChannelAPITokenCreateCalls())
func (mock *PeerChannelsStoreMock) PeerChannelAPITokenCreateCalls() []struct {
	Ctx  context.Context
	Args *payd.PeerChannelAPITokenStoreArgs
} {
	var calls []struct {
		Ctx  context.Context
		Args *payd.PeerChannelAPITokenStoreArgs
	}
	mock.lockPeerChannelAPITokenCreate.RLock()
	calls = mock.calls.PeerChannelAPITokenCreate
	mock.lockPeerChannelAPITokenCreate.RUnlock()
	return calls
}

// PeerChannelAPITokensCreate calls PeerChannelAPITokensCreateFunc.
func (mock *PeerChannelsStoreMock) PeerChannelAPITokensCreate(ctx context.Context, args ...*payd.PeerChannelAPITokenStoreArgs) error {
	if mock.PeerChannelAPITokensCreateFunc == nil {
		panic("PeerChannelsStoreMock.PeerChannelAPITokensCreateFunc: method is nil but PeerChannelsStore.PeerChannelAPITokensCreate was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args []*payd.PeerChannelAPITokenStoreArgs
	}{
		Ctx:  ctx,
		Args: args,
	}
	mock.lockPeerChannelAPITokensCreate.Lock()
	mock.calls.PeerChannelAPITokensCreate = append(mock.calls.PeerChannelAPITokensCreate, callInfo)
	mock.lockPeerChannelAPITokensCreate.Unlock()
	return mock.PeerChannelAPITokensCreateFunc(ctx, args...)
}

// PeerChannelAPITokensCreateCalls gets all the calls that were made to PeerChannelAPITokensCreate.
// Check the length with:
//     len(mockedPeerChannelsStore.PeerChannelAPITokensCreateCalls())
func (mock *PeerChannelsStoreMock) PeerChannelAPITokensCreateCalls() []struct {
	Ctx  context.Context
	Args []*payd.PeerChannelAPITokenStoreArgs
} {
	var calls []struct {
		Ctx  context.Context
		Args []*payd.PeerChannelAPITokenStoreArgs
	}
	mock.lockPeerChannelAPITokensCreate.RLock()
	calls = mock.calls.PeerChannelAPITokensCreate
	mock.lockPeerChannelAPITokensCreate.RUnlock()
	return calls
}

// PeerChannelAccount calls PeerChannelAccountFunc.
func (mock *PeerChannelsStoreMock) PeerChannelAccount(ctx context.Context, args *payd.PeerChannelIDArgs) (*payd.PeerChannelAccount, error) {
	if mock.PeerChannelAccountFunc == nil {
		panic("PeerChannelsStoreMock.PeerChannelAccountFunc: method is nil but PeerChannelsStore.PeerChannelAccount was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args *payd.PeerChannelIDArgs
	}{
		Ctx:  ctx,
		Args: args,
	}
	mock.lockPeerChannelAccount.Lock()
	mock.calls.PeerChannelAccount = append(mock.calls.PeerChannelAccount, callInfo)
	mock.lockPeerChannelAccount.Unlock()
	return mock.PeerChannelAccountFunc(ctx, args)
}

// PeerChannelAccountCalls gets all the calls that were made to PeerChannelAccount.
// Check the length with:
//     len(mockedPeerChannelsStore.PeerChannelAccountCalls())
func (mock *PeerChannelsStoreMock) PeerChannelAccountCalls() []struct {
	Ctx  context.Context
	Args *payd.PeerChannelIDArgs
} {
	var calls []struct {
		Ctx  context.Context
		Args *payd.PeerChannelIDArgs
	}
	mock.lockPeerChannelAccount.RLock()
	calls = mock.calls.PeerChannelAccount
	mock.lockPeerChannelAccount.RUnlock()
	return calls
}

// PeerChannelCloseChannel calls PeerChannelCloseChannelFunc.
func (mock *PeerChannelsStoreMock) PeerChannelCloseChannel(ctx context.Context, channelID string) error {
	if mock.PeerChannelCloseChannelFunc == nil {
		panic("PeerChannelsStoreMock.PeerChannelCloseChannelFunc: method is nil but PeerChannelsStore.PeerChannelCloseChannel was just called")
	}
	callInfo := struct {
		Ctx       context.Context
		ChannelID string
	}{
		Ctx:       ctx,
		ChannelID: channelID,
	}
	mock.lockPeerChannelCloseChannel.Lock()
	mock.calls.PeerChannelCloseChannel = append(mock.calls.PeerChannelCloseChannel, callInfo)
	mock.lockPeerChannelCloseChannel.Unlock()
	return mock.PeerChannelCloseChannelFunc(ctx, channelID)
}

// PeerChannelCloseChannelCalls gets all the calls that were made to PeerChannelCloseChannel.
// Check the length with:
//     len(mockedPeerChannelsStore.PeerChannelCloseChannelCalls())
func (mock *PeerChannelsStoreMock) PeerChannelCloseChannelCalls() []struct {
	Ctx       context.Context
	ChannelID string
} {
	var calls []struct {
		Ctx       context.Context
		ChannelID string
	}
	mock.lockPeerChannelCloseChannel.RLock()
	calls = mock.calls.PeerChannelCloseChannel
	mock.lockPeerChannelCloseChannel.RUnlock()
	return calls
}

// PeerChannelCreate calls PeerChannelCreateFunc.
func (mock *PeerChannelsStoreMock) PeerChannelCreate(ctx context.Context, args *payd.PeerChannelCreateArgs) error {
	if mock.PeerChannelCreateFunc == nil {
		panic("PeerChannelsStoreMock.PeerChannelCreateFunc: method is nil but PeerChannelsStore.PeerChannelCreate was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args *payd.PeerChannelCreateArgs
	}{
		Ctx:  ctx,
		Args: args,
	}
	mock.lockPeerChannelCreate.Lock()
	mock.calls.PeerChannelCreate = append(mock.calls.PeerChannelCreate, callInfo)
	mock.lockPeerChannelCreate.Unlock()
	return mock.PeerChannelCreateFunc(ctx, args)
}

// PeerChannelCreateCalls gets all the calls that were made to PeerChannelCreate.
// Check the length with:
//     len(mockedPeerChannelsStore.PeerChannelCreateCalls())
func (mock *PeerChannelsStoreMock) PeerChannelCreateCalls() []struct {
	Ctx  context.Context
	Args *payd.PeerChannelCreateArgs
} {
	var calls []struct {
		Ctx  context.Context
		Args *payd.PeerChannelCreateArgs
	}
	mock.lockPeerChannelCreate.RLock()
	calls = mock.calls.PeerChannelCreate
	mock.lockPeerChannelCreate.RUnlock()
	return calls
}

// PeerChannelsOpened calls PeerChannelsOpenedFunc.
func (mock *PeerChannelsStoreMock) PeerChannelsOpened(ctx context.Context, channelType payd.PeerChannelHandlerType) ([]payd.PeerChannel, error) {
	if mock.PeerChannelsOpenedFunc == nil {
		panic("PeerChannelsStoreMock.PeerChannelsOpenedFunc: method is nil but PeerChannelsStore.PeerChannelsOpened was just called")
	}
	callInfo := struct {
		Ctx         context.Context
		ChannelType payd.PeerChannelHandlerType
	}{
		Ctx:         ctx,
		ChannelType: channelType,
	}
	mock.lockPeerChannelsOpened.Lock()
	mock.calls.PeerChannelsOpened = append(mock.calls.PeerChannelsOpened, callInfo)
	mock.lockPeerChannelsOpened.Unlock()
	return mock.PeerChannelsOpenedFunc(ctx, channelType)
}

// PeerChannelsOpenedCalls gets all the calls that were made to PeerChannelsOpened.
// Check the length with:
//     len(mockedPeerChannelsStore.PeerChannelsOpenedCalls())
func (mock *PeerChannelsStoreMock) PeerChannelsOpenedCalls() []struct {
	Ctx         context.Context
	ChannelType payd.PeerChannelHandlerType
} {
	var calls []struct {
		Ctx         context.Context
		ChannelType payd.PeerChannelHandlerType
	}
	mock.lockPeerChannelsOpened.RLock()
	calls = mock.calls.PeerChannelsOpened
	mock.lockPeerChannelsOpened.RUnlock()
	return calls
}

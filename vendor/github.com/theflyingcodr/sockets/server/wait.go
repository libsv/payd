package server

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/theflyingcodr/sockets"
)

type waitMessage struct {
	errs chan error
	msg  chan *sockets.Message
}

func newWaitMessage() *waitMessage {
	return &waitMessage{
		errs: make(chan error),
		msg:  make(chan *sockets.Message),
	}
}

func (w *waitMessage) deliver(msg *sockets.Message) {
	w.msg <- msg
}

func (w *waitMessage) wait(ctx context.Context) (*sockets.Message, error) {
	defer func() {
		close(w.errs)
		close(w.msg)
	}()
	for {
		select {
		case msg := <-w.msg:
			return msg, nil
		case err := <-w.errs:
			return nil, err
		case <-ctx.Done():
			return nil, errors.New("timeout waiting for message")
		}
	}
}

// waitMessages is a thread safe map wrapper.
type waitMessages struct {
	sync.RWMutex
	msgs map[string]*waitMessage
}

func newWaitMessgaes() *waitMessages {
	return &waitMessages{
		RWMutex: sync.RWMutex{},
		msgs:    make(map[string]*waitMessage),
	}
}

func (w *waitMessages) message(correlationID string) *waitMessage {
	w.RLock()
	defer w.RUnlock()
	return w.msgs[correlationID]
}

func (w *waitMessages) add(correlationID string, wm *waitMessage) {
	w.Lock()
	defer w.Unlock()
	w.msgs[correlationID] = wm
}

func (w *waitMessages) delete(correlationID string) {
	w.Lock()
	defer w.Unlock()
	delete(w.msgs, correlationID)
}

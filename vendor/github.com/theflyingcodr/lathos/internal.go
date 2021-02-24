package lathos

import (
	"runtime/debug"

	"github.com/google/uuid"
)

// InternalError can be implemented to create errors used
// to capture internal faults. These could then be sent to an
// error logging system to be rectified.
// In terms of a web server, this would be a 5XX error.
type ErrInternal struct {
	id      string
	message string
	stack   string
}

func NewErrInternal(err error) ErrInternal {
	return ErrInternal{
		id:      uuid.New().String(),
		message: err.Error(),
		stack:   string(debug.Stack()),
	}
}

func (e ErrInternal) ID() string {
	return e.id
}

func (e ErrInternal) Message() string {
	return e.message
}

func (e ErrInternal) Stack() string {
	return e.stack
}

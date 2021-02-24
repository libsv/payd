package lathos

import (
	"github.com/pkg/errors"
)

// ClientError defines an error that could be returned to a caller.
// This can be called to build a message type of your choosing to match
// the transport being used by the server such as http, grpc etc
type ClientError interface {
	ID() string
	Code() string
	Title() string
	Detail() string
	error
}

// IsClientError
func IsClientError(err error) bool {
	var t ClientError
	return errors.As(err, &t)
}

// InternalError can be implemented to create errors used
// to capture internal faults. These could then be sent to an
// error logging system to be rectified.
// In terms of a web server, this would be a 5XX error.
type InternalError interface {
	ID() string
	Message() string
	Stack() string
}

// IsClientError
func IsInternalError(err error) bool {
	var t InternalError
	return errors.As(err, &t)
}

type NotFound interface {
	NotFound() bool
}

func IsNotFound(err error) bool {
	var t NotFound
	return errors.As(err, &t)
}

type Duplicate interface {
	Duplicate() bool
}

func IsDuplicate(err error) bool {
	var t Duplicate
	return errors.As(err, &t)
}

type NotAuthorised interface {
	NotAuthorised() bool
}

// IsNotAuthorised will check that and error or it's cause was of the NotAuthorised type.
func IsNotAuthorised(err error) bool {
	var t NotAuthorised
	return errors.As(err, &t)
}

type NotAuthenticated interface {
	NotAuthenticated() bool
}

// IsNotAuthenticated will check that an error is a NotAuthenticated type.
func IsNotAuthenticated(err error) bool {
	var t NotAuthenticated
	return errors.As(err, &t)
}

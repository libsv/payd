package errs

import (
	"github.com/pkg/errors"
)

type NotFound interface {
	NotFound() bool
}

func IsNotFound(err error) bool {
	nf, ok := errors.Cause(err).(NotFound)
	return ok && nf.NotFound()
}

type Duplicate interface {
	IsDuplicate() bool
}

func IsDuplicate(err error) bool {
	nf, ok := errors.Cause(err).(Duplicate)
	return ok && nf.IsDuplicate()
}

type ClientError interface {
	Error() string
	Response() ([]byte, error)
	Headers() (int, map[string]string)
}

type ServerError interface {
	Error() string
	Response() ([]byte, error)
	Headers() (int, map[string]string)
}

type NotFoundError struct {
	Cause  error
	Detail string
	ID     int
}

func NewNotFoundError(cause error, detail string, id int) NotFoundError {
	return NotFoundError{
		Cause:  cause,
		Detail: detail,
		ID:     id,
	}
}

func (m NotFoundError) NotFound() bool {
	return true
}

type DuplicateError struct {
	Cause  error
	Detail string
	ID     int
}

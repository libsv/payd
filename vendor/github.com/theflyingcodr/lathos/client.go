package lathos

import (
	"github.com/google/uuid"
)

// ErrClient can be implemented to create an error
// that can be returned to a user, the intention is to not
// log these errors as client errors could cover validation
// issues, bad inputs etc.
// In terms of a web server this would be a 4XX error.
type ErrClient struct {
	id     string
	code   string
	title  string
	detail string
}

func newErrClient(code, detail string) ErrClient {
	return ErrClient{
		id:     uuid.New().String(),
		code:   code,
		detail: detail,
	}
}

func (e ErrClient) ID() string {
	return e.id
}

func (e ErrClient) Code() string {
	return e.code
}

func (e ErrClient) Title() string {
	return e.title
}

func (e ErrClient) Detail() string {
	return e.detail
}

func (e ErrClient) Error() string {
	return e.title + ": " + e.detail
}

type ErrNotFound struct {
	ErrClient
}

// NewErrNotFound will create and return a new NotFound error.
// You can supply a code which can be set in your application to identify
// a particular error in code such as E404.
// Detail can be supplied to give more context to the error, ie
// "resource 123 does not exist".
func NewErrNotFound(code, detail string) ErrNotFound {
	c := newErrClient(code, detail)
	c.title = "Not found"
	return ErrNotFound{
		ErrClient: c,
	}
}

func (e ErrNotFound) NotFound() bool {
	return true
}

type ErrDuplicate struct {
	ErrClient
}

// NewErrDuplicate will create and return a new Duplicate error.
// You can supply a code which can be set in your application to identify
// a particular error in code such as E404.
// Detail can be supplied to give more context to the error, ie
// "resource 123 already exists".
func NewErrDuplicate(code, detail string) ErrDuplicate {
	c := newErrClient(code, detail)
	c.title = "Item already exists"
	return ErrDuplicate{
		ErrClient: c,
	}
}

func (e ErrDuplicate) Duplicate() bool {
	return true
}

type ErrNotAuthenticated struct {
	ErrClient
}

// NewErrDuplicate will create and return a new Duplicate error.
// You can supply a code which can be set in your application to identify
// a particular error in code such as E404.
// Detail can be supplied to give more context to the error, ie
// "resource 123 already exists".
func NewErrNotAuthenticated(code, detail string) ErrNotAuthenticated {
	c := newErrClient(code, detail)
	c.title = "Item already exists"
	return ErrNotAuthenticated{
		ErrClient: c,
	}
}

func (e ErrNotAuthenticated) NotAuthenticated() bool {
	return true
}

type ErrNotAuthorised struct {
	ErrClient
}

// NewErrDuplicate will create and return a new Duplicate error.
// You can supply a code which can be set in your application to identify
// a particular error in code such as E404.
// Detail can be supplied to give more context to the error, ie
// "resource 123 already exists".
func NewErrNotAuthorised(code, detail string) ErrNotAuthorised {
	c := newErrClient(code, detail)
	c.title = "Permission denied"
	return ErrNotAuthorised{
		ErrClient: c,
	}
}

func (e ErrNotAuthorised) NotAuthorised() bool {
	return true
}

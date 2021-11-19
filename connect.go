package payd

import (
	"context"

	validator "github.com/theflyingcodr/govalidator"
)

// ConnectArgs identify the invoiceId / channelID to connect to.
type ConnectArgs struct {
	InvoiceID string `param:"invoiceId"`
}

// Validate will check that invoice arguments match expectations.
func (c *ConnectArgs) Validate() error {
	return validator.New().
		Validate("invoiceID", validator.StrLength(c.InvoiceID, 1, 30)).
		Err()
}

// ConnectService is used to connect this wallet to an async channel server, this could be sockets, peer channels
// or something else.
type ConnectService interface {
	Connect(ctx context.Context, args ConnectArgs) error
}

// ConnectWriter handles data writes when creating a new async connection.
type ConnectWriter interface {
	Connect(ctx context.Context, args ConnectArgs) error
}

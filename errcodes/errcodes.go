package errcodes

// error codes used throughout application.
const (
	ErrDuplicatePayment = "D1"
	ErrExpiredPayment   = "E1"

	ErrInvoiceNotFound          = "N0001"
	ErrInvoicesNotFound         = "N0002"
	ErrDestinationsNotFound     = "N0003"
	ErrDestinationsFailedCreate = "N0004"
	ErrTxNotFound               = "N0005"
)

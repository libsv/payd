package gopayd

import (
	"context"
)

type P2PTransactionArgs struct {
	Alias     string
	Domain    string
	PaymentID string
	TxHex     string
}

type P2PTransaction struct {
	TxHex    string
	Metadata P2PTransactionMetadata
}

type P2PTransactionMetadata struct {
	Note      string // A human readable bit of information about the payment
	PubKey    string // Public key to validate the signature (if signature is given)
	Sender    string // The paymail of the person that originated the transaction
	Signature string
}

type P2PCapabilityArgs struct {
	Domain string
	BrfcID string
}

type P2POutputCreateArgs struct {
	Domain string
	Alias  string
}

type P2PPayment struct {
	Satoshis uint64
}

type PaymailReader interface {
	Capability(ctx context.Context, args P2PCapabilityArgs) (string, error)
}

type PaymailWriter interface {
	OutputsCreate(ctx context.Context, args P2POutputCreateArgs, req P2PPayment) ([]*Output, error)
	Broadcast(ctx context.Context, args P2PTransactionArgs, req P2PTransaction) error
}

type PaymailReaderWriter interface {
	PaymailReader
	PaymailWriter
}

package gopayd

import (
	"context"
	"time"

	"gopkg.in/guregu/null.v3"
)

// TxoCreate is used when creating outputs as part of a paymentRequest
// the script keys created are stored and later picked up
// and validated when the user sends a payment.
//
// These are partial txos and will be further hydrated when a transaction
// is sent spending them.
type TxoCreate struct {
	TxID           string      `db:"txid"`
	TxHex          string      `db:"txhex"`
	Vout           uint64      `db:"vout"`
	PaymentID      null.String `db:"paymentID"`
	KeyName        string      `db:"keyname"`
	Satoshis       uint64      `db:"satoshis"`
	LockingScript  string      `db:"lockingscript"`
	DerivationPath string      `db:"derivationpath"`
}

// PartialTxoCreate is used when creating outputs as part of a paymentRequest
// the script keys created are stored and later picked up
// and validated when the user sends a payment.
//
// These are partial txos and will be further hydrated when a transaction
// is sent spending them.
type PartialTxoCreate struct {
	PaymentID      null.String
	KeyName        string
	DerivationPath string
	LockingScript  string
	Satoshis       uint64
}

// UnspentTxoArgs are used to located an unfulfilled txo.
type UnspentTxoArgs struct {
	Keyname       string `db:"keyname" json:"account"`
	LockingScript string `db:"lockingscript" json:"-"`
	Satoshis      uint64 `db:"satoshis" json:"-"`
}

// UnspentTxo is an unfulfilled txo not yet linked to a transaction.
type UnspentTxo struct {
	KeyName        string
	DerivationPath string
	LockingScript  string
	Satoshis       uint64
	CreatedAt      time.Time
	ModifiedAt     time.Time
}

// TxoWriter is used to add transaction information to a data store.
type TxoWriter interface {
	// Txo will return a txo a list a txos.
	StoreChange(ctx context.Context, args TxoCreate) error
	// PartialTxoCreate will add a partial txo to a data store.
	PartialTxoCreate(ctx context.Context, req PartialTxoCreate) error
	// PartialTxosCreate will add an array of partial txos to a data store.
	PartialTxosCreate(ctx context.Context, req []*PartialTxoCreate) error
}

// TxoReader is used to read tx information from a data store.
type TxoReader interface {
	// UnspentTxos will return all unspent txos.
	UnspentTxos(ctx context.Context, req UnspentTxoArgs) ([]Txo, error)
	// Txo will return a txo a list a txos.
	ReserveTxos(ctx context.Context, args TxoReserveArgs) ([]Txo, error)
	// PartialTxo will return a txo that has not yet been assigned to a transaction.
	PartialTxo(ctx context.Context, args UnspentTxoArgs) (*UnspentTxo, error)
	// PartialTxo will return a txo that has not yet been assigned to a transaction.
	PartialTxoByPaymentID(ctx context.Context, args InvoiceArgs) ([]UnspentTxo, error)
	DerivationPath(ctx context.Context, ls string) (string, error)
}

type TxoReaderWriter interface {
	TxoReader
	TxoWriter
}

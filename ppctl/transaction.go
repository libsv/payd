package ppctl

import (
	"context"
	"time"
)

type Transaction struct {
	PaymentID string
	TxID      string
	TxHex     string
	CreatedAt time.Time
	Outputs   []Txo
}

type Txo struct {
	Outpoint       string
	TxID           string
	Vout           int64
	KeyName        string
	DerivationPath string
	LockingScript  string
	Satoshis       int64
	SpentAt        time.Time
	SpendingTxID   string
	CreatedAt      time.Time
	ModifiedAt     time.Time
}

type CreateTxo struct {
	Outpoint       string
	TxID           string
	Vout           int64
	KeyName        string
	DerivationPath string
	LockingScript  string
	Satoshis       int64
}

// SpendTxo can be used to update a transaction out with information
// on when it was spent and by what transaction.
type SpendTxo struct {
	SpentAt      *time.Time
	SpendingTxID string
}

// SpendTxoArgs are used to identify the transaction output to mark as spent.
type SpendTxoArgs struct {
	Outpoint string
}

type TransactionStore interface {
	// Create can be implemented to store a Transaction in a datastore.
	Create(ctx context.Context, req Transaction) (*Transaction, error)
	// Spend can be used to mark a transaction as spent.
	Spend(ctx context.Context, args SpendTxoArgs, req SpendTxo) (*Transaction, error)
}

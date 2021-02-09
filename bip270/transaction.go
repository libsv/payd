package bip270

import (
	"context"
	"time"

	"github.com/libsv/go-bt"
)

type Txo struct {
	Outpoint       string
	Instance       time.Time
	TxID           string
	Vout           int64
	Alias          string
	DerivationPath string
	ScriptPubKey   string
	Satoshis       int64
	ReservedAt     time.Time
	SpentAt        time.Time
	SpendingTxID   string
	CreatedAt      time.Time
	ModifiedAt     time.Time
}

type CreateTxo struct {
	Outpoint       string
	Instance       time.Time
	TxID           string
	Vout           int64
	Alias          string
	DerivationPath string
	ScriptPubKey   string
	Satoshis       int64
	ReservedAt     time.Time
	SpentAt        *time.Time
	SpendingTxID   string
}

// Tx represents a stored transaction.
type Tx struct {
	TxID      string
	TxHex     string
	Txos      []Txo
	CreatedAt time.Time
}

// CreateTxArgs contains all arguments required to create transactions.
type CreateTxArgs struct {
	DerivationPath string
}

type TransactionStore interface {
	// Create can be implemented to store a Transaction in a datastore.
	Create(ctx context.Context, args CreateTxArgs, req *bt.Tx) (*Tx, error)
}

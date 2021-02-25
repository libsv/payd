package gopayd

import (
	"time"

	"gopkg.in/guregu/null.v3"
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
	Satoshis       uint64
	SpentAt        null.Time
	SpendingTxID   null.String
	CreatedAt      time.Time
	ModifiedAt     time.Time
}

type CreateTransaction struct {
	PaymentID string      `db:"paymentId"`
	TxID      string      `db:"txId"`
	TxHex     string      `db:"txHex"`
	Outputs   []CreateTxo `db:"-"`
}

type CreateTxo struct {
	Outpoint       string `db:"outpoint"`
	TxID           string `db:"txId"`
	Vout           int    `db:"vout"`
	KeyName        string `db:"keyname"`
	DerivationPath string `db:"derivationPath"`
	LockingScript  string `db:"lockingScript"`
	Satoshis       uint64 `db:"satoshis"`
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

type TxoArgs struct {
	Outpoint string
}

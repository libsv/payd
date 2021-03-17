package gopayd

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

type Transaction struct {
	PaymentID string    `db:"paymentID"`
	TxID      string    `db:"txid"`
	TxHex     string    `db:"txhex"`
	CreatedAt time.Time `db:"createdAt"`
	Outputs   []Txo     `db:"-"`
}

type Txo struct {
	Outpoint       string      `db:"outpoint"`
	TxID           string      `db:"txid"`
	Vout           int         `db:"vout"`
	KeyName        null.String `db:"keyname"`
	DerivationPath null.String `db:"derivationpath"`
	LockingScript  string      `db:"lockingscript"`
	Satoshis       uint64      `db:"satoshis"`
	SpentAt        null.Time   `db:"spentat"`
	SpendingTxID   null.String `db:"spendingtxid"`
	CreatedAt      time.Time   `db:"createdAt"`
	ModifiedAt     time.Time   `db:"modifiedAt"`
}

type CreateTransaction struct {
	PaymentID string      `db:"paymentID"`
	TxID      string      `db:"txid"`
	TxHex     string      `db:"txhex"`
	Outputs   []CreateTxo `db:"-"`
}

type CreateTxo struct {
	Outpoint       string      `db:"outpoint"`
	TxID           string      `db:"txid"`
	Vout           int         `db:"vout"`
	KeyName        null.String `db:"keyname"`
	DerivationPath null.String `db:"derivationpath"`
	LockingScript  string      `db:"lockingscript"`
	Satoshis       uint64      `db:"satoshis"`
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

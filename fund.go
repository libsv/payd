package gopayd

import (
	"gopkg.in/guregu/null.v3"
)

// TxoReserveArgs the args for getting funds.
type TxoReserveArgs struct {
	Account     string
	Offset      int
	Limit       int
	ReservedFor string
}

// TxoStoreRequest the request for storing a fund in the db.
type TxoStoreRequest struct {
	TxID      string      `db:"txid"`
	TxHex     string      `db:"txhex"`
	PaymentID null.String `db:"paymentID"`
	Txos      []*Txo
}

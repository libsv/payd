package payd

import (
	"context"
)

// TxoCreate will add utxos to our data store linked by a destinationId.
// These are added when a user submit a tx to pay an invoice.
type TxoCreate struct {
	Outpoint      string `db:"outpoint"`
	DestinationID uint64 `db:"destination_id"`
	TxID          string `db:"tx_id"`
	Vout          uint64 `db:"vout"`
}

// UTXO an internal utxo.
type UTXO struct {
	Outpoint       string `db:"outpoint"`
	TxID           string `db:"tx_id"`
	Vout           uint32 `db:"vout"`
	Satoshis       uint64 `db:"satoshis"`
	LockingScript  string `db:"locking_script"`
	DerivationPath string `db:"derivation_path"`
}

// UTXOReserve takes args for marking a utxo in the db as reserved.
type UTXOReserve struct {
	ReservedFor string
	Satoshis    uint64
}

// UTXOUnreserve takes args for unreserving reserved utxos in the db.
type UTXOUnreserve struct {
	ReservedFor string
}

// UTXOSpend takes args for marking a utxo in the db as spent.
type UTXOSpend struct {
	SpendingTxID string `db:"spending_txid"`
	Reservation  string `db:"reserved_for"`
}

// TxoWriter is used to add transaction information to a data store.
type TxoWriter interface {
	// TxosCreate will add an array of txos to a data store.
	// TxosCreate(ctx context.Context, req []*TxoCreate) error
	UTXOReserve(ctx context.Context, req UTXOReserve) ([]UTXO, error)
	UTXOUnreserve(ctx context.Context, req UTXOUnreserve) error
	UTXOSpend(ctx context.Context, req UTXOSpend) error
}

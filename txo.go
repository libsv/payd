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

// TxoWriter is used to add transaction information to a data store.
type TxoWriter interface {
	// TxosCreate will add an array of txos to a data store.
	TxosCreate(ctx context.Context, req []*TxoCreate) error
}

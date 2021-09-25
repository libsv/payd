package payd

import (
	"context"
	"time"

	"gopkg.in/guregu/null.v3"
)

// defines states a transaction can have.
const (
	StateTxBroadcast TxState = "broadcast"
	StateTxFailed    TxState = "failed"
	StateTxPending   TxState = "pending"
)

// TxState defines states a transaction can have.
type TxState string

// Transaction defines a single transaction.
type Transaction struct {
	PaymentID string    `db:"paymentid"`
	TxID      string    `db:"tx_id"`
	TxHex     string    `db:"tx_hex"`
	CreatedAt time.Time `db:"created_at"`
	Outputs   []Txo     `db:"-"`
	State     string    `enums:"pending,broadcast,failed,deleted"`
}

// Txo defines a single txo and can be returned from the data store.
type Txo struct {
	Outpoint       string      `db:"outpoint"`
	TxID           string      `db:"tx_id"`
	Vout           int         `db:"vout"`
	KeyName        null.String `db:"key_name"`
	DerivationPath null.String `db:"derivation_path"`
	LockingScript  string      `db:"locking_script"`
	Satoshis       uint64      `db:"satoshis"`
	SpentAt        null.Time   `db:"spent_at"`
	SpendingTxID   null.String `db:"spending_txid"`
	CreatedAt      time.Time   `db:"created_at"`
	ModifiedAt     time.Time   `db:"updated_at"`
}

// TransactionCreate is used to insert a tx into the data store.
// To save calls, Txos can be included to also add in the same transaction.
type TransactionCreate struct {
	InvoiceID null.String  `db:"invoice_id"`
	TxID      string       `db:"tx_id"`
	TxHex     string       `db:"tx_hex"`
	Outputs   []*TxoCreate `db:"-"`
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

// TxoArgs is used to get a single txo.
type TxoArgs struct {
	Outpoint string
}

// TransactionArgs are used to identify a specific tx.
type TransactionArgs struct {
	TxID string `db:"tx_id"`
}

// TransactionStateUpdate contains information to update a tx.
type TransactionStateUpdate struct {
	State TxState `db:"state"`
}

// TransactionWriter will add and update transaction data.
type TransactionWriter interface {
	TransactionCreate(ctx context.Context, req TransactionCreate) error
	// TransactionUpdateState can be used to change a tx state (failed, broadcast).
	TransactionUpdateState(ctx context.Context, args TransactionArgs, req TransactionStateUpdate) error
	TransactionChangeCreate(ctx context.Context, txArgs TransactionCreate, dArgs DestinationCreate) error
}

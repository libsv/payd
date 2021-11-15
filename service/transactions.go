package service

import (
	"context"
	"fmt"
	"time"

	"github.com/libsv/go-bt/v2"

	"github.com/libsv/payd"
)

type transactions struct {
	transacter payd.Transacter
	destRdr    payd.DestinationsReaderWriter
	tWtr       payd.TransactionWriter
	invRdr     payd.InvoiceReaderWriter
}

// NewTransactions will setup and return a new Transaction service.
func NewTransactions(transacter payd.Transacter, destRdr payd.DestinationsReaderWriter, tWtr payd.TransactionWriter, invRdr payd.InvoiceReaderWriter) *transactions {
	return &transactions{
		transacter: transacter,
		destRdr:    destRdr,
		tWtr:       tWtr,
		invRdr:     invRdr,
	}
}

// Submit will add a finalised tx to the data store.
func (t *transactions) Submit(ctx context.Context, args payd.TransactionSubmitArgs, req payd.TransactionSubmit) error {
	ctx = t.transacter.WithTx(ctx)
	defer func() {
		_ = t.transacter.Rollback(ctx)
	}()
	inv, err := t.invRdr.Invoice(ctx, payd.InvoiceArgs{InvoiceID: args.InvoiceID})
	if err != nil {
		return err
	}
	tx, err := bt.NewTxFromString(req.TxHex)
	if err != nil {
		return err
	}
	txID := tx.TxID()
	destLookup := map[string]uint64{}
	destinations, err := t.destRdr.Destinations(ctx, payd.DestinationsArgs{InvoiceID: args.InvoiceID})
	if err != nil {
		return err
	}
	for _, d := range destinations {
		destLookup[d.LockingScript.String()] = 0
	}

	for i, out := range tx.Outputs {
		if _, ok := destLookup[out.LockingScript.String()]; !ok {
			continue
		}
		destLookup[out.LockingScript.String()] = uint64(i)
	}

	txos := make([]*payd.TxoCreate, len(destinations))
	for i := 0; i < len(destinations); i++ {
		outpoint := destLookup[destinations[i].LockingScript.String()]
		txos[i] = &payd.TxoCreate{
			Outpoint:      fmt.Sprintf("%s%d", txID, outpoint),
			DestinationID: destinations[i].ID,
			TxID:          txID,
			Vout:          outpoint,
		}
	}
	if err := t.tWtr.TransactionCreate(ctx, payd.TransactionCreate{
		InvoiceID: inv.ID,
		TxID:      tx.TxID(),
		TxHex:     req.TxHex,
		Outputs:   txos,
	}); err != nil {
		return err
	}
	// mark tx as broadcast
	if err := t.tWtr.TransactionUpdateState(ctx, payd.TransactionArgs{TxID: tx.TxID()}, payd.TransactionStateUpdate{State: payd.StateTxBroadcast}); err != nil {
		return err
	}

	if _, err := t.invRdr.InvoiceUpdate(ctx, payd.InvoiceUpdateArgs{InvoiceID: args.InvoiceID}, payd.InvoiceUpdatePaid{
		PaymentReceivedAt: time.Now().UTC(),
		RefundTo:          "",
	}); err != nil {
		return err
	}

	return t.transacter.Commit(ctx)
}

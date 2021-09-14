package sqlite

import (
	"context"

	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
)

func (s *sqliteStore) StoreUtxos(ctx context.Context, req gopayd.CreateTransaction) (*gopayd.Transaction, error) {
	tx, err := s.newTx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transaction when inserting transaction to db")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	resp, err := s.txCreateTransaction(tx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transaction and utxos")
	}
	return resp, errors.Wrapf(commit(ctx, tx),
		"failed to commit transaction when adding tx and outputs for paymentID %s", req.PaymentID)
}

// txCreateTransaction takes a db object / transaction and adds a transaction to the data store
// along with utxos, returning the transaction.
// This method can be used with other methods in the store allowing
// multiple methods to be ran in the same db transaction.
func (s *sqliteStore) txCreateTransaction(tx db, req gopayd.CreateTransaction) (*gopayd.Transaction, error) {
	// insert tx and utxos
	if err := handleNamedExec(tx, sqlTransactionCreate, req); err != nil {
		return nil, errors.Wrap(err, "failed to insert new transaction")
	}
	if err := handleNamedExec(tx, sqlTxoUpdate, req.Outputs); err != nil {
		return nil, errors.Wrap(err, "failed to insert transaction outputs")
	}
	var outTx gopayd.Transaction
	if err := tx.Get(&outTx, sqlTransactionByID, req.TxID); err != nil {
		return nil, errors.Wrapf(err, "failed to get stored transaction for paymentID %s", req.PaymentID)
	}
	var outTxos []gopayd.Txo
	if err := tx.Select(&outTxos, sqlTxosByTxID, req.TxID); err != nil {
		return nil, errors.Wrapf(err, "failed to get stored transaction outputs for paymentID %s", req.PaymentID)
	}
	outTx.Outputs = outTxos
	return &outTx, nil
}

func (s *sqliteStore) StoreFund(ctx context.Context, req gopayd.StoreFundRequest) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start transaction when inserting transaction to db")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()
	if err = handleNamedExec(tx, sqlTransactionCreate, req); err != nil {
		return err
	}
	for _, txo := range req.Txos {
		if err = handleNamedExec(tx, sqlTxoCreateAsFund, txo); err != nil {
			return err
		}
	}
	return errors.Wrap(commit(ctx, tx), "failed to commit tx when adding fund")
}

func (s *sqliteStore) SpendFunds(ctx context.Context, req *gopayd.FundsSpendReq, args gopayd.FundsSpendArgs) error {
	tx, err := s.newTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start transaction when spending fund")
	}
	defer func() {
		_ = rollback(ctx, tx)
	}()

	for _, fund := range req.Txos {
		if err = handleNamedExec(tx, sqlFundSpend, struct {
			TxID         string `db:"txid"`
			Vout         int    `db:"vout"`
			KeyName      string `db:"keyname"`
			SpendingTxID string `db:"spendingTxId"`
		}{
			TxID:         fund.TxID,
			Vout:         fund.Vout,
			KeyName:      "client",
			SpendingTxID: req.SpendingTxID,
		}); err != nil {
			return errors.Wrap(err, "failed to set fund to spent")
		}
	}
	return errors.Wrap(commit(ctx, tx), "failed to commit tx when updating funds")
}

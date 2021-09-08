package service

import (
	"context"
	"fmt"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/client"
	"github.com/libsv/payd/errcodes"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos/errs"
)

type payment struct {
	ec    spv.EnvelopeCreator
	pc    client.PaymentCreator
	fSvc  client.FundService
	pkSvc gopayd.PrivateKeyService
	fRwr  client.FundReaderWriter
}

// NewPayment returns a new service for payments.
func NewPayment(ec spv.EnvelopeCreator, pc client.PaymentCreator, fSvc client.FundService, pkSvc gopayd.PrivateKeyService, fRwr client.FundReaderWriter) client.PaymentService {
	return &payment{
		ec:    ec,
		pc:    pc,
		fSvc:  fSvc,
		pkSvc: pkSvc,
		fRwr:  fRwr,
	}
}

func (p *payment) CreatePayment(ctx context.Context, req client.CreatePayment) (*gopayd.PaymentACK, error) {
	epk, err := p.pkSvc.PrivateKey(ctx, "client")
	if err != nil {
		return nil, err
	}
	pk, err := epk.ECPrivKey()
	if err != nil {
		return nil, errors.Wrap(err, "error getting bec private key")
	}

	invoice, err := p.pc.Invoice(ctx, req.ServerURL, gopayd.InvoiceCreate{
		Satoshis: req.Satoshis,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating invoice")
	}

	payReq, err := p.pc.RequestPayment(ctx, req.ServerURL, gopayd.PaymentRequestArgs{
		PaymentID: invoice.PaymentID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error requesting payment")
	}

	tx := bt.NewTx()
	var totalOutput uint64 = 0
	for _, output := range payReq.Outputs {
		script, err := bscript.NewFromHexString(output.Script)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing hex script %s", output.Script)
		}
		if err = tx.AddP2PKHOutputFromScript(script, output.Amount); err != nil {
			return nil, errors.Wrap(err, "error adding output from script")
		}

		totalOutput += output.Amount
	}

	funds, err := p.fRwr.Funds(ctx, client.FundArgs{KeyName: "client"})
	if err != nil {
		return nil, err
	}

	var totalInput uint64 = 0
	for i := 0; i < len(funds) && totalInput < totalOutput; i++ {
		f := funds[i]

		if err = tx.From(f.TxID, uint32(f.Vout), f.LockingScript, f.Satoshis); err != nil {
			return nil, errors.Wrap(err, "err adding input to tx")
		}

		totalInput += f.Satoshis
	}

	if tx.InputCount() < len(funds) {
		funds = append(funds[:tx.InputCount()], funds[tx.InputCount()+1:]...)
	}

	if totalInput < totalOutput {
		return nil, errs.NewErrUnprocessable(errcodes.ErrInsufficientFunds, fmt.Sprintf("insufficient funds: %d", totalInput))
	}

	script, err := bscript.NewP2PKHFromPubKeyEC(pk.PubKey())
	if err != nil {
		return nil, err
	}
	if err = tx.Change(script, bt.NewFeeQuote()); err != nil {
		return nil, err
	}

	n, err := tx.SignAuto(ctx, &bt.LocalSigner{PrivateKey: pk})
	if err != nil {
		return nil, errors.Wrap(err, "error signing tx")
	}
	if len(n) == 0 {
		return nil, errs.NewErrUnprocessable(errcodes.ErrSignInputs, "could not sign tx inputs")
	}

	spv, err := p.ec.CreateEnvelope(ctx, tx)
	if err != nil {
		return nil, err
	}

	ack, err := p.pc.SendPayment(ctx, payReq.PaymentURL, gopayd.CreatePayment{
		Transaction:  tx.String(),
		Memo:         payReq.Memo,
		MerchantData: *payReq.MerchantData,
		SPVEnvelope:  spv,
	})
	if err != nil {
		return nil, errors.Wrap(errs.NewErrUnprocessable(errcodes.ErrPaymentRejected, "payment rejected"), err.Error())
	}

	txID := tx.TxID()
	for _, f := range funds {
		f.SpendingTxID = txID
	}

	if err := p.fRwr.FundsSpend(ctx, funds); err != nil {
		return nil, errors.Wrap(err, "error marking fund as spent")
	}

	if _, err := p.fSvc.FundsCreate(ctx, tx); err != nil {
		return nil, errors.Wrap(err, "error writing change fund")
	}

	return ack, nil
}

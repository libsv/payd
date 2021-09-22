package service

import (
	"context"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/payd/cli/models"
)

type paymentSvc struct {
	ps    models.PaymentStore
	txSig models.Signer
	spvb  spv.EnvelopeCreator
}

// NewPaymentService returns a new payment service.
func NewPaymentService(ps models.PaymentStore, txSig models.Signer, spvb spv.EnvelopeCreator) models.PaymentService {
	return &paymentSvc{
		ps:    ps,
		txSig: txSig,
		spvb:  spvb,
	}
}

func (p *paymentSvc) Request(ctx context.Context, args models.PaymentRequestArgs) (*models.PaymentRequest, error) {
	return p.ps.Request(ctx, args)
}

func (p *paymentSvc) Send(ctx context.Context, args models.PaymentSendArgs) (*models.PaymentAck, error) {
	var signedTx *bt.Tx
	if args.Tx == "" {
		tx := bt.NewTx()
		var totalOutputs uint64
		for _, o := range args.PaymentRequest.Outputs {
			script, err := bscript.NewFromHexString(o.Script)
			if err != nil {
				return nil, err
			}
			if err = tx.AddP2PKHOutputFromScript(script, o.Amount); err != nil {
				return nil, err
			}

			totalOutputs += o.Amount
		}

		signedTxResp, err := p.txSig.FundAndSign(ctx, models.FundAndSignTxRequest{})

		//signedTxResp, err := p.txSig.FundAndSign(ctx, gopayd.FundAndSignTxRequest{
		//	TxHex:     tx.String(),
		//	Account:   args.Account,
		//	PaymentID: args.PaymentRequest.MerchantData.PaymentReference,
		//	Fee:       gopayd.Fee(args.PaymentRequest.Fee),
		//})
		if err != nil {
			return nil, err
		}

		signedTx, err = bt.NewTxFromString(signedTxResp.SignedTx)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		signedTx, err = bt.NewTxFromString(args.Tx)
		if err != nil {
			return nil, err
		}
	}

	spvEnvelope, err := p.spvb.CreateEnvelope(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	pAck, err := p.ps.Submit(ctx, models.PaymentSendArgs{
		Account:        args.Account,
		Transaction:    signedTx.String(),
		PaymentRequest: args.PaymentRequest,
		MerchantData:   args.PaymentRequest.MerchantData,
		Memo:           args.PaymentRequest.Memo,
		SPVEnvelope:    *spvEnvelope,
	})
	if err != nil {
		return nil, err
	}

	return pAck, nil
}

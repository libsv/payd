package service

import (
	"context"

	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/cli/models"
)

type paymentSvc struct {
	ps    models.PaymentStore
	fSvc  models.FundService
	txSig models.Signer
	spvb  spv.EnvelopeCreator
}

// NewPaymentService returns a new payment service.
func NewPaymentService(ps models.PaymentStore, fSvc models.FundService, txSig models.Signer, spvb spv.EnvelopeCreator) models.PaymentService {
	return &paymentSvc{
		ps:    ps,
		fSvc:  fSvc,
		txSig: txSig,
		spvb:  spvb,
	}
}

func (p *paymentSvc) Request(ctx context.Context, args models.PaymentRequestArgs) (*models.PaymentRequest, error) {
	return p.ps.Request(ctx, args)
}

func (p *paymentSvc) Send(ctx context.Context, args models.PaymentSendArgs) (*models.PaymentAck, error) {
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

	signedTxResp, err := p.txSig.FundAndSign(ctx, gopayd.FundAndSignTxRequest{
		TxHex:   tx.String(),
		Account: "client",
		Fee:     gopayd.Fee(args.PaymentRequest.Fee),
	})
	if err != nil {
		return nil, err
	}

	signedTx, err := bt.NewTxFromString(signedTxResp.SignedTx)
	if err != nil {
		return nil, err
	}

	spvEnvelope, err := p.spvb.CreateEnvelope(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	pAck, err := p.ps.Submit(ctx, models.PaymentSendArgs{
		Transaction:    signedTxResp.SignedTx,
		PaymentRequest: args.PaymentRequest,
		MerchantData:   args.PaymentRequest.MerchantData,
		Memo:           args.PaymentRequest.Memo,
		SPVEnvelope:    *spvEnvelope,
	})
	if err != nil {
		return nil, err
	}

	if err := p.fSvc.Spend(ctx, models.FundSpendArgs{
		SpendingTx: signedTxResp.SignedTx,
		Account:    "client",
	}); err != nil {
		return nil, err
	}

	if _, err := p.fSvc.Add(ctx, models.FundAddArgs{
		TxHex:   signedTxResp.SignedTx,
		Account: "client",
	}); err != nil {
		return nil, err
	}

	return pAck, nil
}

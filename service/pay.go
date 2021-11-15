package service

import (
	"context"
	"fmt"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/libsv/payd/data/http"
)

type pay struct {
	storeTx payd.Transacter
	p4      http.P4
	spvc    payd.EnvelopeService
	svrCfg  *config.Server
}

// NewPayService returns a pay service.
func NewPayService(storeTx payd.Transacter, p4 http.P4, spvc payd.EnvelopeService, svrCfg *config.Server) payd.PayService {
	return &pay{
		storeTx: storeTx,
		p4:      p4,
		spvc:    spvc,
		svrCfg:  svrCfg,
	}
}

type derivationSigner struct {
	pathMap       map[*bscript.Script]string
	masterPrivKey *bip32.ExtendedKey
}

// Signer returns a signer configured for a provided *bscript.Script.
func (l derivationSigner) Signer(ctx context.Context, script *bscript.Script) (bt.Signer, error) {
	path, ok := l.pathMap[script]
	if !ok {
		return nil, errors.New("derivation path does not exist for script")
	}
	extPrivKey, err := l.masterPrivKey.DeriveChildFromPath(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to derive child from path %s for locking script %s", path, script.String())
	}

	privKey, err := extPrivKey.ECPrivKey()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ec private key for script %s", script.String())
	}

	return &bt.LocalSigner{
		PrivateKey: privKey,
	}, nil
}

// Pay takes a pay-to url and performs a payment procedure, ultimately sending money to the
// url.
func (p *pay) Pay(ctx context.Context, req payd.PayRequest) (*payd.PaymentACK, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Retrieve the payment request information from the receiver.
	payReq, err := p.p4.PaymentRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to request payment for url %s", req.PayToURL)
	}
	// begin a transaction that can be picked up by other services etc for rollbacks on failure.
	ctx = p.storeTx.WithTx(ctx)
	defer func() {
		_ = p.storeTx.Rollback(ctx)
	}()
	env, err := p.spvc.Envelope(ctx, payd.EnvelopeArgs{PayToURL: req.PayToURL}, *payReq)
	if err != nil {
		return nil, errors.Wrapf(err, "envelope creation failed for '%s'", req.PayToURL)
	}
	// Send the payment to the p4 server.
	ack, err := p.p4.PaymentSend(ctx, req, payd.PaymentSend{
		SPVEnvelope: env,
		ProofCallbacks: map[string]payd.ProofCallback{
			"https://" + p.svrCfg.Hostname + "/api/v1/proofs/" + env.TxID: {},
		},
		MerchantData: payd.User{
			Name:         payReq.MerchantData.Name,
			Email:        payReq.MerchantData.Email,
			Avatar:       payReq.MerchantData.Avatar,
			Address:      payReq.MerchantData.Address,
			ExtendedData: payReq.MerchantData.ExtendedData,
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send payment %s", req.PayToURL)
	}
	if ack.Error > 0 {
		return nil, fmt.Errorf("failed to send payment. Code '%d' reason '%s'", ack.Error, ack.Memo)
	}
	if err := p.storeTx.Commit(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to commit tx")
	}
	return ack, nil
}

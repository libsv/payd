package service

import (
	"context"
	"fmt"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-dpp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/libsv/payd/data/http"
	lerrs "github.com/theflyingcodr/lathos/errs"
)

type pay struct {
	storeTx    payd.Transacter
	txWtr      payd.TransactionWriter
	dpp        http.DPP
	spvc       payd.EnvelopeService
	pcStr      payd.PeerChannelsStore
	pcNotifSvc payd.PeerChannelsNotifyService
	svrCfg     *config.Server
}

// NewPayService returns a pay service.
func NewPayService(storeTx payd.Transacter, dpp http.DPP, spvc payd.EnvelopeService, svrCfg *config.Server, pcNotifSvc payd.PeerChannelsNotifyService, pcStr payd.PeerChannelsStore, txWtr payd.TransactionWriter) payd.PayService {
	return &pay{
		storeTx:    storeTx,
		txWtr:      txWtr,
		dpp:        dpp,
		spvc:       spvc,
		svrCfg:     svrCfg,
		pcStr:      pcStr,
		pcNotifSvc: pcNotifSvc,
	}
}

type derivationSigner struct {
	pathMap       map[*bscript.Script]string
	masterPrivKey *bip32.ExtendedKey
}

// Unlocker returns a signer configured for a provided *bscript.Script.
func (l derivationSigner) Unlocker(ctx context.Context, script *bscript.Script) (bt.Unlocker, error) {
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

	return &bt.LocalUnlocker{
		PrivateKey: privKey,
	}, nil
}

// Pay takes a pay-to url and performs a payment procedure, ultimately sending money to the
// url.
func (p *pay) Pay(ctx context.Context, req payd.PayRequest) (*dpp.PaymentACK, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Retrieve the payment request information from the receiver.
	payReq, err := p.dpp.PaymentRequest(ctx, req)
	if err != nil {
		if errors.As(err, &lerrs.ErrUnprocessable{}) {
			return nil, lerrs.NewErrUnprocessable("U002", "failed to request payment for url "+req.PayToURL+" : "+err.Error())
		}

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
	// Send the payment to the dpp proxy server.
	ack, err := p.dpp.PaymentSend(ctx, req, dpp.Payment{
		Ancestors: env,
		ProofCallbacks: map[string]dpp.ProofCallback{
			"https://" + p.svrCfg.Hostname + "/api/v1/proofs/" + env.TxID: {},
		},
		MerchantData: dpp.Merchant{
			Name:         payReq.MerchantData.Name,
			Email:        payReq.MerchantData.Email,
			AvatarURL:    payReq.MerchantData.AvatarURL,
			Address:      payReq.MerchantData.Address,
			ExtendedData: payReq.MerchantData.ExtendedData,
		},
	})
	if err != nil {
		if err := p.txWtr.TransactionUpdateState(ctx, payd.TransactionArgs{TxID: env.TxID}, payd.TransactionStateUpdate{State: payd.StateTxFailed}); err != nil {
			log.Error().Err(errors.Wrap(err, "failed to update tx after failed broadcast"))
		}
		return nil, errors.Wrapf(err, "failed to send payment %s", req.PayToURL)
	}
	if ack.Error > 0 {
		return nil, fmt.Errorf("failed to send payment. Code '%d' reason '%s'", ack.Error, ack.Memo)
	}

	// Update tx state to broadcast
	// Just logging errors here as I don't want to roll back tx now tx is broadcast.
	if err := p.txWtr.TransactionUpdateState(ctx, payd.TransactionArgs{TxID: env.TxID}, payd.TransactionStateUpdate{State: payd.StateTxBroadcast}); err != nil {
		log.Error().Err(errors.Wrap(err, "failed to update tx to broadcast state"))
	}

	if ack.PeerChannel == nil {
		return ack, nil
	}

	if err := p.pcStr.PeerChannelCreate(ctx, &payd.PeerChannelCreateArgs{
		PeerChannelAccountID: 0,
		ChannelID:            ack.PeerChannel.ChannelID,
		ChannelHost:          ack.PeerChannel.Host,
		ChannelPath:          ack.PeerChannel.Path,
		ChannelType:          payd.PeerChannelHandlerTypeProof,
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to store channel %s/%s in db", ack.PeerChannel.Host, ack.PeerChannel.ChannelID)
	}
	if err := p.pcStr.PeerChannelAPITokenCreate(ctx, &payd.PeerChannelAPITokenStoreArgs{
		Token:                 ack.PeerChannel.Token,
		CanRead:               true,
		CanWrite:              false,
		PeerChannelsChannelID: ack.PeerChannel.ChannelID,
		Role:                  "notification",
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to store token %s", ack.PeerChannel.Token)
	}

	if err := p.pcNotifSvc.Subscribe(ctx, &payd.PeerChannel{
		ID:    ack.PeerChannel.ChannelID,
		Token: ack.PeerChannel.Token,
		Host:  ack.PeerChannel.Host,
		Path:  ack.PeerChannel.Path,
		Type:  payd.PeerChannelHandlerTypeProof,
	}); err != nil {
		log.Err(err)
	}
	if err := p.storeTx.Commit(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to commit tx")
	}
	return ack, nil
}

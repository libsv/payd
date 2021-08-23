package ppctl

import (
	"context"
	"math"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"

	gopayd "github.com/libsv/payd"

	"github.com/libsv/payd/config"
)

const (
	keyname              = "keyname"
	derivationPathPrefix = "0"
	duplicatePayment     = "D0001"
)

type mapiOutputs struct {
	privKeySvc gopayd.PrivateKeyService
	store      gopayd.PaymentRequestReaderWriter
	txrunner   gopayd.Transacter
}

// NewMapiOutputs will create and return a new payment service.
func NewMapiOutputs(env *config.Server, privKeySvc gopayd.PrivateKeyService, txrunner gopayd.Transacter, store gopayd.PaymentRequestReaderWriter) *mapiOutputs {
	if env == nil || env.Hostname == "" {
		log.Fatal("env hostname should be set")
	}
	return &mapiOutputs{privKeySvc: privKeySvc, store: store, txrunner: txrunner}
}

// CreatePaymentRequest handles setting up a new PaymentRequest response and can use and optional existing paymentID.
//
// This will split the requested satoshis into denominations, with each denomintation getting
// its own locking script to help with privacy when payments are broadcast.
// This is limited however, for full privacy you'd probably want a new TX per script.
func (p *mapiOutputs) CreateOutputs(ctx context.Context, args gopayd.OutputsCreate) ([]*gopayd.Output, error) {
	ctx = p.txrunner.WithTx(ctx)
	// get our master key
	priv, err := p.privKeySvc.PrivateKey(ctx, keyname)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// get the current derivation path counter, we will increment this
	// to generate a new deterministic derivationPath from the master.
	counter, err := p.store.DerivationCounter(ctx, gopayd.DerivationCounterArgs{Key: keyname})
	if err != nil {
		return nil, errors.Wrap(err, "failed to check payment request is a duplicate")
	}
	totOutputs := math.Ceil(float64(args.Satoshis) / float64(args.Denomination))
	if err := p.store.IncrementKeyCounter(ctx, gopayd.DerivationIncrementArgs{
		Key:    keyname,
		Offset: uint64(totOutputs),
	}); err != nil {
		return nil, errors.Wrap(err, "failed to increment derivation path")
	}
	txos := make([]*gopayd.Txo, 0)
	for c := counter + 1; c <= uint64(totOutputs); c++ {
		path := bip32.DerivePath(counter)
		key, err := p.privKeySvc.DeriveChildFromKey(priv, path)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to derive key when creating output")
		}
		pubKey, err := key.ECPubKey()
		if err != nil {
			return nil, errors.WithMessage(err, "failed to derive key when creating output")
		}
		s, err := bscript.NewP2PKHFromPubKeyBytes(pubKey.SerialiseCompressed())
		if err != nil {
			return nil, errors.WithMessage(err, "failed to derive key when creating output")
		}
		sats := args.Denomination * c
		if sats > args.Satoshis {
			sats = sats - args.Satoshis
		} else {
			sats = args.Denomination
		}
		txos = append(txos, &gopayd.Txo{
			KeyName:        null.StringFrom(keyname),
			DerivationPath: null.StringFrom(path),
			LockingScript:  s.String(),
			Satoshis:       sats,
			CreatedAt:      time.Now().UTC(),
			ModifiedAt:     time.Now().UTC(),
		})

	}
	if err := p.txrunner.Commit(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to created payment")
	}
	return outs, nil
}

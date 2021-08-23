package ppctl

import (
	"context"
	"math"
	"math/rand"

	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"

	gopayd "github.com/libsv/payd"

	"github.com/libsv/payd/config"
)

const (
	keyname              = "keyname"
	derivationPathPrefix = "0"
	duplicatePayment     = "D0001"
)

type mapiOutputs struct {
	privKeySvc    gopayd.PrivateKeyService
	txoWtr        gopayd.TxoWriter
	derivationRdr gopayd.DerivationReader
}

// NewMapiOutputs will create and return a new payment service.
func NewMapiOutputs(env *config.Server, privKeySvc gopayd.PrivateKeyService, txoWtr gopayd.TxoWriter, derivationRdr gopayd.DerivationReader) *mapiOutputs {
	if env == nil || env.Hostname == "" {
		log.Fatal("env hostname should be set")
	}
	return &mapiOutputs{privKeySvc: privKeySvc, derivationRdr: derivationRdr, txoWtr: txoWtr}
}

// CreatePaymentRequest handles setting up a new PaymentRequest response and can use and optional existing paymentID.
//
// This will split the requested satoshis into denominations, with each denomintation getting
// its own locking script to help with privacy when payments are broadcast.
// This is limited however, for full privacy you'd probably want a new TX per script.
func (p *mapiOutputs) CreateOutputs(ctx context.Context, args gopayd.OutputsCreate) ([]*gopayd.Output, error) {
	// get our master key
	priv, err := p.privKeySvc.PrivateKey(ctx, keyname)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	totOutputs := math.Ceil(float64(args.Satoshis) / float64(args.Denomination))
	txos := make([]*gopayd.TxoCreate, totOutputs, totOutputs)
	oo := make([]*gopayd.Output, totOutputs, totOutputs)
	for i := 0; i < int(totOutputs); i++ {
		var path string
		for { // attempt to create a unique derivation path
			seed := rand.Uint64()
			path = bip32.DerivePath(seed)
			exists, err := p.derivationRdr.DerivationPathExists(ctx, gopayd.DerivationExistsArgs{
				KeyName: keyname,
				Path:    path,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to check derivation path exists when creating new payment request output")
			}
			if !exists {
				break
			}
		}
		pubKey, err := priv.DerivePublicKeyFromPath(path)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new extended key when creating new payment request output")
		}
		s, err := bscript.NewP2PKHFromPubKeyBytes(pubKey)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to derive key when creating output")
		}
		sats := args.Denomination * uint64(i+1)
		if sats > args.Satoshis {
			sats = sats - args.Satoshis
		} else {
			sats = args.Denomination
		}
		txos = append(txos, &gopayd.TxoCreate{
			KeyName:        keyname,
			DerivationPath: path,
			LockingScript:  s.String(),
			Satoshis:       sats,
		})
		oo = append(oo, &gopayd.Output{
			Amount: sats,
			Script: s.String(),
		})
	}
	if err := p.txoWtr.TxosCreate(ctx, txos); err != nil {
		return nil, errors.Wrap(err, "failed to store outputs")
	}
	return oo, nil
}

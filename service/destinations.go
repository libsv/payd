package service

import (
	"context"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/payd/config"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/libsv/payd"
)

const (
	// TODO - this will need changed.
	keyname = "masterkey"
)

type destinations struct {
	deployCfg  *config.Wallet
	privKeySvc payd.PrivateKeyService
	destRdrWtr payd.DestinationsReaderWriter
	derivRdr   payd.DerivationReader
	invRdr     payd.InvoiceReader
	seed       payd.SeedService
}

// NewDestinationsService will setup and return a new Output Service for creating and reading payment destination info.
func NewDestinationsService(deployCfg *config.Wallet, privKeySvc payd.PrivateKeyService, destRdrWtr payd.DestinationsReaderWriter, derivRdr payd.DerivationReader, invRdr payd.InvoiceReader, seed payd.SeedService) *destinations {
	return &destinations{
		deployCfg:  deployCfg,
		privKeySvc: privKeySvc,
		destRdrWtr: destRdrWtr,
		derivRdr:   derivRdr,
		invRdr:     invRdr,
		seed:       seed,
	}
}

// Create will split satoshis into multiple denominations and store
// as denominations waiting to be fulfilled in a tx.
func (d *destinations) DestinationsCreate(ctx context.Context, req payd.DestinationsCreate) (*payd.Destination, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	// Get the master key associated with the beneficiary of the invoice.
	key := "masterkey"
	priv, err := d.privKeySvc.PrivateKey(ctx, key, req.UserID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// TODO - split requested satsohis in some way
	// 1 for now - we may decide to increase or split output in future so
	// keeping the code here flexible
	totOutputs := 1
	destinations := make([]payd.DestinationCreate, 0, totOutputs)
	for i := 0; i < totOutputs; i++ {
		// TODO - run in a go routine when we start splitting
		var path string
		for { // attempt to create a unique derivation path
			seed, err := d.seed.Uint64()
			if err != nil {
				return nil, errors.Wrap(err, "failed to create seed for derivation path")
			}
			path = bip32.DerivePath(seed)
			exists, err := d.derivRdr.DerivationPathExists(ctx, payd.DerivationExistsArgs{
				KeyName: key,
				UserID:  req.UserID,
				Path:    path,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to check derivation path exists when creating new destination")
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
		// use the below if we decide to split outputs
		/*sats := args.Denomination * uint64(i+1)
		if sats > args.Satoshis {
			sats = sats - args.Satoshis
		} else {
			sats = args.Denomination
		}*/
		destinations = append(destinations, payd.DestinationCreate{
			UserID:         req.UserID,
			DerivationPath: path,
			Script:         s.String(),
			Satoshis:       req.Satoshis,
		})
	}
	oo, err := d.destRdrWtr.DestinationsCreate(ctx, payd.DestinationsCreateArgs{InvoiceID: req.InvoiceID}, destinations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store destinations")
	}
	return &payd.Destination{
		Outputs: oo,
	}, nil
}

// Destinations given the args, will return a set of Destinations.
func (d *destinations) Destinations(ctx context.Context, args payd.DestinationsArgs) (*payd.Destination, error) {
	if err := args.Validate(); err != nil {
		return nil, err
	}

	var invoice *payd.Invoice
	g := new(errgroup.Group)
	g.Go(func() error {
		i, err := d.invRdr.Invoice(ctx, payd.InvoiceArgs{InvoiceID: args.InvoiceID})
		if err != nil {
			return errors.Wrapf(err, "failed to get invoice for invoiceID '%s' when getting destinations", args.InvoiceID)
		}
		invoice = i
		return nil
	})
	var outputs []payd.Output
	g.Go(func() error {
		oo, err := d.destRdrWtr.Destinations(ctx, args)
		if err != nil {
			return errors.Wrapf(err, "failed to read destinations for invoiceID '%s'", args.InvoiceID)
		}
		outputs = oo
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, errors.WithStack(err)
	}
	return &payd.Destination{
		Network:     string(d.deployCfg.Network),
		SPVRequired: invoice.SPVRequired,
		Outputs:     outputs,
		CreatedAt:   invoice.CreatedAt,
		ExpiresAt:   invoice.ExpiresAt.ValueOrZero(),
	}, nil
}

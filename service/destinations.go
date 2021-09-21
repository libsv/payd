package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"

	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos/errs"

	gopayd "github.com/libsv/payd"
)

const (
	// TODO - this will need changed
	keyname = "masterkey"
)

type destinations struct {
	privKeySvc gopayd.PrivateKeyService
	destWtr    gopayd.DestinationsWriter
	derivRdr   gopayd.DerivationReader
	feeRdr     gopayd.FeeReader
}

// NewDestinationsService will setup and return a new Output Service for creating and reading payment destination info.
func NewDestinationsService(privKeySvc gopayd.PrivateKeyService, destWtr gopayd.DestinationsWriter, derivRdr gopayd.DerivationReader, feeRdr gopayd.FeeReader) *destinations {
	return &destinations{
		privKeySvc: privKeySvc,
		destWtr:    destWtr,
		derivRdr:   derivRdr,
		feeRdr:     feeRdr,
	}
}

// Create will split satoshis into multiple denominations and store
// as denominations waiting to be fulfilled in a tx.
func (d *destinations) DestinationsCreate(ctx context.Context, req gopayd.DestinationsCreate) (*gopayd.Destination, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	// get our master private key
	priv, err := d.privKeySvc.PrivateKey(ctx, keyname)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// TODO - split requested satsohis in some way
	// 1 for now - we may decide to increase or split output in future so
	// keeping the code here flexible
	totOutputs := 1
	destinations := make([]gopayd.DestinationCreate, 0, totOutputs)
	for i := 0; i < totOutputs; i++ {
		// TODO - run in a go routine when we start splitting
		var path string
		for { // attempt to create a unique derivation path
			seed, err := randUint64()
			if err != nil {
				return nil, errors.New("failed to create seed for derivation path")
			}
			path = bip32.DerivePath(seed)
			exists, err := d.derivRdr.DerivationPathExists(ctx, gopayd.DerivationExistsArgs{
				KeyName: keyname,
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
		destinations = append(destinations, gopayd.DestinationCreate{
			Keyname:        keyname,
			DerivationPath: path,
			Script:         s.String(),
			Satoshis:       req.Satoshis,
		})
	}
	oo, err := d.destWtr.DestinationsCreate(ctx, gopayd.DestinationsCreateArgs{InvoiceID: req.InvoiceID}, destinations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store destinations")
	}
	// GET Fees
	fees, err := d.feeRdr.Fees(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get fees when creating destinations")
	}
	return &gopayd.Destination{
		Outputs: oo,
		Fees:    fees,
	}, nil
}

// Destinations given the args, will return a set of Destinations.
func (d *destinations) Destinations(ctx context.Context, args gopayd.DestinationsArgs) (*gopayd.Destination, error) {
	// TODO - implement this
	return nil, errs.NewErrUnprocessable("U01", "not implemented")
}

func randUint64() (uint64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}

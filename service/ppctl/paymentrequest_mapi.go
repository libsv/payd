package ppctl

import (
	"context"
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/pkg/errors"
	"gopkg.in/guregu/null.v3"

	gopayd "github.com/libsv/payd"

	"github.com/theflyingcodr/lathos/errs"

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
func (p *mapiOutputs) CreateOutputs(ctx context.Context, args gopayd.OutputsCreate) ([]*gopayd.Output, error) {
	ctx = p.txrunner.WithTx(ctx)
	exists, err := p.store.DerivationPathExists(ctx, gopayd.DerivationPathExistsArgs{PaymentID: "1"})
	if err != nil {
		return nil, errors.Wrap(err, "failed to check payment request is a duplicate")
	}
	if exists {
		return nil, errs.NewErrDuplicate(
			duplicatePayment, fmt.Sprintf("payment request for paymentID %s already exists", "1"))
	}
	// get the master key stored
	// TODO: later we will allow users to provide their own key for now we've hardcoded to keyname
	xprv, err := p.privKeySvc.PrivateKey(ctx, keyname)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	outs := make([]*gopayd.Output, 0)
	// generate a new child for each output
	// TODO: figure out how many outputs we need?
	// TODO: what should derivation path be, prefix is just hardcoded for now, this could be a user setting.
	dp, err := p.store.DerivationPathCreate(ctx, gopayd.DerivationPathCreate{
		PaymentID: "1",
		Prefix:    derivationPathPrefix,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create derivationPath when creating payment request")
	}
	// create output from key and derivation path
	o, err := p.generateOutput(xprv, dp.Path, args.Satoshis)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	outs = append(outs, &gopayd.Output{
		Amount: o.Satoshis,
		Script: o.LockingScriptHexString(),
	})

	// store outputs so we can get them later for validation
	if err := p.storeKeys(ctx, keyname, dp.ID, outs); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := p.txrunner.Commit(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to created payment")
	}
	return outs, nil
}

func (p *mapiOutputs) generateOutput(xprv *bip32.ExtendedKey, derivPath string, satoshis uint64) (*bt.Output, error) {
	key, err := p.privKeySvc.DeriveChildFromKey(xprv, derivPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pubKey, err := p.privKeySvc.PubFromXPrv(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	o, err := bscript.NewP2PKHFromPubKeyBytes(pubKey.SerialiseCompressed())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &bt.Output{
		Satoshis:      satoshis,
		LockingScript: o,
	}, nil
}

// storeKeys will store each key along with keyname and derivation path
// to allow us to validate the outputs sent in the users payment.
// If there is a failure all will be rolled back.
func (p *mapiOutputs) storeKeys(ctx context.Context, keyName string, derivID int, outs []*gopayd.Output) error {
	keys := make([]gopayd.CreateScriptKey, 0)
	for _, o := range outs {
		keys = append(keys, gopayd.CreateScriptKey{
			LockingScript: o.Script,
			KeyName:       null.StringFrom(keyName),
			DerivationID:  null.IntFrom(int64(derivID)),
		})
	}
	return errors.Wrap(p.store.CreateScriptKeys(ctx, keys), "failed to create payment request when storing key map")
}

package ppctl

import (
	"context"
	"fmt"
	"time"

	"github.com/bitcoinsv/bsvutil/hdkeychain"
	"github.com/labstack/gommon/log"
	"github.com/libsv/go-bt"
	gopayd "github.com/libsv/payd"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"

	"github.com/libsv/payd/config"
	"github.com/libsv/payd/ipaymail"
	"github.com/theflyingcodr/lathos"
)

const (
	keyname              = "keyname"
	derivationPathPrefix = "0"
	duplicatePayment     = "D0001"
)

type paymentRequestService struct {
	env        *config.Server
	wallet     *config.Wallet
	privKeySvc gopayd.PrivateKeyService
	store      gopayd.PaymentRequestReaderWriter
	txrunner   gopayd.Transacter
}

// NewPaymentRequestService will create and return a new payment service.
func NewPaymentRequestService(env *config.Server, wallet *config.Wallet, privKeySvc gopayd.PrivateKeyService, txrunner gopayd.Transacter, store gopayd.PaymentRequestReaderWriter) *paymentRequestService {
	if env == nil || env.Hostname == "" {
		log.Fatal("env hostname should be set")
	}
	return &paymentRequestService{env: env, wallet: wallet, privKeySvc: privKeySvc, store: store, txrunner: txrunner}
}

// CreatePaymentRequest handles setting up a new PaymentRequest response and can use and optional existing paymentID.
func (p *paymentRequestService) CreatePaymentRequest(ctx context.Context, args gopayd.PaymentRequestArgs) (*gopayd.PaymentRequest, error) {
	if err := validator.New().
		Validate("paymentID", validator.NotEmpty(args.PaymentID)).
		Validate("hostname", validator.NotEmpty(p.env)); err.Err() != nil {
		return nil, err
	}
	inv, err := p.store.Invoice(ctx, gopayd.InvoiceArgs{PaymentID: args.PaymentID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get invoice when creating payment request")
	}
	// TODO: I hate this here
	ctx = p.txrunner.WithTx(ctx)
	exists, err := p.store.DerivationPathExists(ctx, gopayd.DerivationPathExistsArgs{PaymentID: args.PaymentID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to check payment request is a duplicate")
	}
	if exists {
		return nil, lathos.NewErrDuplicate(
			duplicatePayment, fmt.Sprintf("payment request for paymentID %s already exists", args.PaymentID))
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
		PaymentID: args.PaymentID,
		Prefix:    derivationPathPrefix,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create derivationPath when creating payment request")
	}
	// create output from from key and derivation path
	o, err := p.generateOutput(xprv, dp.Path, inv.Satoshis)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	outs = append(outs, &gopayd.Output{
		Amount: o.Satoshis,
		Script: o.GetLockingScriptHexString(),
	})
	// store outputs so we can get them later for validation
	if err := p.storeKeys(ctx, keyname, dp.ID, outs); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := p.txrunner.Commit(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to created payment")
	}
	return &gopayd.PaymentRequest{
		Network:             p.wallet.Network,
		Outputs:             outs,
		CreationTimestamp:   time.Now().UTC().Unix(),
		ExpirationTimestamp: time.Now().Add(24 * time.Hour).UTC().Unix(),
		PaymentURL:          fmt.Sprintf("http://%s/v1/payment/%s", p.env.Hostname, args.PaymentID),
		Memo:                fmt.Sprintf("Payment request for invoice %s", args.PaymentID),
		MerchantData: &gopayd.MerchantData{
			AvatarURL:    p.wallet.MerchantAvatarURL,
			MerchantName: p.wallet.MerchantName,
		},
	}, nil
}

func (p *paymentRequestService) generateOutput(xprv *hdkeychain.ExtendedKey, derivPath string, satoshis uint64) (*bt.Output, error) {
	key, err := p.privKeySvc.DeriveChildFromKey(xprv, derivPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pubKey, err := p.privKeySvc.PubFromXPrv(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	o, err := bt.NewP2PKHOutputFromPubKeyBytes(pubKey, satoshis)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return o, nil
}

// createPaymailOutputs is not currently used but will be when we incorporate this feature.
func (p *paymentRequestService) createPaymailOutputs(paymentID string, outs []*gopayd.Output) ([]*gopayd.Output, error) {
	ref, os, err := ipaymail.GetP2POutputs("jad@moneybutton.com", 10000)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get paymail outputs")
	}
	log.Debugf("reference: %s", ref)

	ipaymail.ReferencesMap[paymentID] = ref

	// change returned hexString output script into bytes TODO: understand what i wrote
	for _, o := range os {
		out := &gopayd.Output{
			Amount: o.Satoshis,
			Script: o.Script,
		}
		outs = append(outs, out)
	}
	return outs, nil
}

// storeKeys will store each key along with keyname and derivation path
// to allow us to validate the outputs sent in the users payment.
// If there is a failure all will be rolled back.
func (p *paymentRequestService) storeKeys(ctx context.Context, keyName string, derivID int, outs []*gopayd.Output) error {
	keys := make([]gopayd.CreateScriptKey, 0)
	for _, o := range outs {
		keys = append(keys, gopayd.CreateScriptKey{
			LockingScript: o.Script,
			KeyName:       keyName,
			DerivationID:  derivID,
		})
	}
	return errors.Wrap(p.store.CreateScriptKeys(ctx, keys), "failed to create payment request when storing key map")
}

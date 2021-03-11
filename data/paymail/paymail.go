package paymail

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos"
	gopaymail "github.com/tonicpow/go-paymail"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/data/paymail/models"
)

// TODO: test this code.
// powmailClient is a wrapper interface for tonicpow.go-paymail
// to allow easier unit testing of this code.
type powmailClient interface {
	GetSRVRecord(service, protocol, domainName string) (srv *net.SRV, err error)
	GetCapabilities(target string, port int) (response *gopaymail.Capabilities, err error)
	GetP2PPaymentDestination(p2pURL, alias, domain string, paymentRequest *gopaymail.PaymentRequest) (response *gopaymail.PaymentDestination, err error)
	SendP2PTransaction(p2pURL, alias, domain string, transaction *gopaymail.P2PTransaction) (response *gopaymail.P2PTransactionResponse, err error)
	VerifyPubKey(verifyURL, alias, domain, pubKey string) (response *gopaymail.Verification, err error)
}

type paymail struct {
	cli    powmailClient
	cstore map[string]*gopaymail.Capabilities
}

// NewPaymail will setup and return a new paymail data store used
// to create and send paymail transactions.
func NewPaymail(cli powmailClient) *paymail {
	return &paymail{
		cli:    cli,
		cstore: map[string]*gopaymail.Capabilities{},
	}
}

// Capability will return a capability or a notfound error if it could not be found.
func (p *paymail) Capability(ctx context.Context, args gopayd.P2PCapabilityArgs) (string, error) {
	c, ok := p.cstore[args.Domain]
	if ok && c.Has(args.BrfcID, "") {
		return c.GetString(args.BrfcID, ""), nil
	}
	srv, err := p.cli.GetSRVRecord(gopaymail.DefaultServiceName, gopaymail.DefaultProtocol, args.Domain)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get srv record for %s", args.Domain)
	}
	cp, err := p.cli.GetCapabilities(srv.Target, int(srv.Port))
	if err != nil {
		return "", errors.Wrapf(err, "failed to get capabilities for %s", args.Domain)
	}
	p.cstore[args.Domain] = cp
	if cp.Has(args.BrfcID, "") {
		return cp.GetString(args.BrfcID, ""), nil
	}
	return "", lathos.NewErrNotFound("N001",
		fmt.Sprintf("brfcID [%s] not found for domain [%s]", args.BrfcID, args.Domain))
}

// OutputsCreate will create outputs for the provided payment information. Args are used to gather capability information
// a lathos.NotFound error may be returned if the paymail or brfc doesn't exist.
func (p *paymail) OutputsCreate(ctx context.Context, args gopayd.P2POutputCreateArgs, req gopayd.P2PPayment) ([]*gopayd.Output, error) {
	url, err := p.Capability(ctx, gopayd.P2PCapabilityArgs{
		Domain: args.Domain,
		BrfcID: gopaymail.BRFCP2PPaymentDestination,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get BRFCP2P Payment Destination for domain %s", args.Domain)
	}
	resp, err := p.cli.GetP2PPaymentDestination(url, args.Alias, args.Domain, &gopaymail.PaymentRequest{Satoshis: req.Satoshis})
	if err != nil {
		if err.Error() == "paymail address not found" {
			return nil, lathos.NewErrNotFound("N003", err.Error())
		}
		return nil, errors.Wrapf(err, "failed to generate paymail outputs for alias %s", args.Alias)
	}
	return models.OutputsToPayd(resp.Outputs), nil
}

// Broadcast will transmit a transaction via paymail to a destination address.
// A lathos.NotFound error will be returned if the paymail destination doesn't exist or
// the paymail service doesn't have BRFCP2PTransactions capability.
func (p *paymail) Broadcast(ctx context.Context, args gopayd.P2PTransactionArgs, req gopayd.P2PTransaction) error {
	url, err := p.Capability(ctx, gopayd.P2PCapabilityArgs{
		Domain: args.Domain,
		BrfcID: gopaymail.BRFCP2PTransactions,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to send transaction for paymentID %s", args.PaymentID)
	}
	if req.Metadata.PubKey != "" {
		if err := p.verifyPubKey(ctx, args.Alias, args.Domain, req.Metadata.PubKey); err != nil {
			return errors.WithStack(err)
		}
	}
	if _, err := p.cli.SendP2PTransaction(url, args.Alias, args.Domain, &gopaymail.P2PTransaction{
		Hex:       req.TxHex,
		Reference: args.PaymentID,
		MetaData: &gopaymail.P2PMetaData{
			Note:      req.Metadata.Note,
			PubKey:    req.Metadata.PubKey,
			Sender:    req.Metadata.Sender,
			Signature: req.Metadata.Signature,
		},
	}); err != nil {
		if err.Error() == "paymail address not found" {
			return lathos.NewErrNotFound("N002", err.Error())
		}
		return errors.Wrapf(err, "failed to send transaction for paymentID %s", args.PaymentID)
	}
	return nil
}

func (p *paymail) verifyPubKey(ctx context.Context, alias, domain, pubKey string) error {
	url, err := p.Capability(ctx, gopayd.P2PCapabilityArgs{
		Domain: domain,
		BrfcID: gopaymail.BRFCVerifyPublicKeyOwner,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to send transaction for paymentID %s", args.PaymentID)
	}
	v, err := p.cli.VerifyPubKey(url, alias, domain, pubKey)
	if err != nil {
		return errors.Wrap(err, "failed to validate public key")
	}
	if v.Match {
		return nil
	}
	return errors.New("public key did not match handle")
}

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

type powmailClient interface {
	GetSRVRecord(service, protocol, domainName string) (srv *net.SRV, err error)
	GetCapabilities(target string, port int) (response *gopaymail.Capabilities, err error)
	GetP2PPaymentDestination(p2pURL, alias, domain string, paymentRequest *gopaymail.PaymentRequest) (response *gopaymail.PaymentDestination, err error)
	SendP2PTransaction(p2pURL, alias, domain string, transaction *gopaymail.P2PTransaction) (response *gopaymail.P2PTransactionResponse, err error)
}

type paymail struct {
	cli powmailClient
	cap map[string]*gopaymail.Capabilities
}

func NewPaymail(cli powmailClient) *paymail {
	return &paymail{cli: cli}
}

// Capability will return a capability or a notfound error if it could not be found.
func (p *paymail) Capability(ctx context.Context, args gopayd.P2PCapabilityArgs) (string, error) {
	c, ok := p.cap[args.Domain]
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
	p.cap[args.Domain] = cp
	if cp.Has(args.BrfcID, "") {
		return cp.GetString(args.BrfcID, ""), nil
	}
	return "", lathos.NewErrNotFound("N001",
		fmt.Sprintf("brfcID [%s] not found for domain [%s]", args.BrfcID, args.Domain))
}

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

func (p *paymail) Broadcast(ctx context.Context, args gopayd.P2PTransactionArgs, req gopayd.P2PTransaction) error {
	url, err := p.Capability(ctx, gopayd.P2PCapabilityArgs{
		Domain: args.Domain,
		BrfcID: gopaymail.BRFCP2PTransactions,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to send transaction for paymentID %s", args.PaymentID)
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

package mapi

import (
	"context"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/config"
	"github.com/pkg/errors"
	"github.com/tonicpow/go-minercraft"
)

type broadcast struct {
	client *minercraft.Client
	cfg    *config.MApi
}

// NewBroadcast will setup and return a new MAPI broadcast data store.
func NewBroadcast(cfg *config.MApi, client *minercraft.Client) *broadcast {
	return &broadcast{client: client, cfg: cfg}
}

// Broadcast will submit a transaction to mapi for inclusion in a block.
// Any errors will be returned, no error denotes success.
func (b *broadcast) Send(ctx context.Context, args gopayd.CreatePaymentArgs, req gopayd.CreatePayment) error {
	// TODO: Support callback url for notifications.
	resp, err := b.client.SubmitTransaction(
		b.client.MinerByName(b.cfg.MinerName),
		&minercraft.Transaction{RawTx: req.Transaction})
	if err != nil {
		return errors.Wrap(err, "failed to submit transaction to minerpool")
	}
	if resp.Results.ReturnResult == minercraft.QueryTransactionSuccess {
		return nil
	}
	return errors.Errorf("failed to submit transaction %s", resp.Results.ResultDescription)
}

package mapi

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tonicpow/go-minercraft"

	gopayd "github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

type broadcast struct {
	client *minercraft.Client
	cfg    *config.MApi
	svrCfg *config.Server
}

// NewBroadcast will setup and return a new MAPI broadcast data store.
func NewBroadcast(cfg *config.MApi, svrCfg *config.Server, client *minercraft.Client) *broadcast {
	return &broadcast{client: client, cfg: cfg, svrCfg: svrCfg}
}

// Broadcast will submit a transaction to mapi for inclusion in a block.
// Any errors will be returned, no error denotes success.
func (b *broadcast) Send(ctx context.Context, args gopayd.SendTransactionArgs, req gopayd.CreatePayment) error {
	resp, err := b.client.SubmitTransaction(
		b.client.MinerByName(b.cfg.MinerName),
		&minercraft.Transaction{
			RawTx:              req.Transaction,
			CallBackURL:        b.svrCfg.Hostname + "/api/v1/proofs/" + args.TxID,
			CallBackToken:      "",
			MerkleFormat:       "TSC",
			CallBackEncryption: "",
			MerkleProof:        true,
			DsCheck:            true,
		})
	if err != nil {
		return errors.Wrap(err, "failed to submit transaction to minerpool")
	}
	if resp.Results.ReturnResult == minercraft.QueryTransactionSuccess {
		return nil
	}
	return errors.Errorf("failed to submit transaction %s", resp.Results.ResultDescription)
}

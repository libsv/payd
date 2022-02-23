package mapi

import (
	"context"
	"time"

	"github.com/libsv/go-bt/v2"
	"github.com/pkg/errors"
	"github.com/tonicpow/go-minercraft"

	"github.com/libsv/payd"
	"github.com/libsv/payd/config"
)

type minercraftMapi struct {
	client *minercraft.Client
	cfg    *config.MApi
	fq     *bt.FeeQuote
}

// NewMapi will setup and return a new MAPI minercraftMapi data store.
func NewMapi(cfg *config.MApi, client *minercraft.Client) *minercraftMapi {
	return &minercraftMapi{client: client, cfg: cfg, fq: bt.NewFeeQuote()}
}

// Broadcast will submit a transaction to mapi for inclusion in a block.
// Any errors will be returned, no error denotes success.
func (m *minercraftMapi) Broadcast(ctx context.Context, args payd.BroadcastArgs, tx *bt.Tx) error {
	resp, err := m.client.SubmitTransaction(ctx,
		m.client.MinerByName(m.cfg.MinerName),
		&minercraft.Transaction{
			RawTx:              tx.String(),
			CallBackURL:        args.CallbackURL,
			CallBackToken:      args.Token,
			MerkleFormat:       minercraft.MerkleFormatTSC,
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
	if resp.Results.ResultDescription == "Transaction already in the mempool" {
		// This is a hack for paymail where both parties broadcast the transaction from their own end.
		// What ends up happening is one beats the other to the punch.
		// If the transaction is already in the mempool then the status is a success with regards the intention of the request.
		// Despite the fact that the miner's mapi will say 500 Error.
		// Although it's not clear whether we will still get the merkleproofs or not.
		// This should be fixed in MAPI not here, long term.
		return nil
	}
	return errors.Errorf("failed to submit transaction %s", resp.Results.ResultDescription)
}

// FeeQuote will return the current fees for the configured miner. If the fee has not
// expired we will return the current memoized fee quote.
func (m *minercraftMapi) FeeQuote(ctx context.Context) (*bt.FeeQuote, error) {
	if !m.fq.Expired() {
		return m.fq, nil
	}
	fq, err := m.client.FeeQuote(ctx, m.client.MinerByName(m.cfg.MinerName))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read fees for ")
	}
	if !fq.Validated {
		return m.fq, nil
	}
	stdfee := fq.Quote.GetFee(string(bt.FeeTypeStandard))
	datafee := fq.Quote.GetFee(string(bt.FeeTypeData))
	m.fq.AddQuote(bt.FeeTypeStandard, &bt.Fee{
		FeeType: bt.FeeTypeStandard,
		MiningFee: bt.FeeUnit{
			Satoshis: stdfee.MiningFee.Satoshis,
			Bytes:    stdfee.MiningFee.Bytes,
		},
		RelayFee: bt.FeeUnit{
			Satoshis: stdfee.RelayFee.Satoshis,
			Bytes:    stdfee.RelayFee.Bytes,
		},
	})
	m.fq.AddQuote(bt.FeeTypeData, &bt.Fee{
		FeeType: bt.FeeTypeData,
		MiningFee: bt.FeeUnit{
			Satoshis: datafee.MiningFee.Satoshis,
			Bytes:    datafee.MiningFee.Bytes,
		},
		RelayFee: bt.FeeUnit{
			Satoshis: datafee.RelayFee.Satoshis,
			Bytes:    datafee.RelayFee.Bytes,
		},
	})
	exp, err := time.Parse(time.RFC3339, fq.Quote.ExpirationTime)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse expiration time when getting fee quote")
	}
	m.fq.UpdateExpiry(exp.UTC())
	return m.fq, nil
}

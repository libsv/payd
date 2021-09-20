package gopayd

import (
	"context"
	"errors"

	"github.com/libsv/go-bt/v2"
	validator "github.com/theflyingcodr/govalidator"
)

// SignerService interfaces a signing service.
type SignerService interface {
	FundAndSignTx(ctx context.Context, req FundAndSignTxRequest) (*SignTxResponse, error)
}

// FundAndSignTxRequest the request for signing and funding a tx.
type FundAndSignTxRequest struct {
	PaymentID string `json:"paymentId"`
	TxHex     string `json:"tx"`
	Account   string `json:"account"`
	Fee       Fee    `json:"fee"`
}

// Validate validates.
func (f *FundAndSignTxRequest) Validate() error {
	v := validator.New().
		Validate("account", validator.NotEmpty(f.Account)).
		Validate("fee.standard.miningFee.satsohis", validator.MinInt(f.Fee.Standard.MiningFee.Satoshis, 0)).
		Validate("fee.standard.miningFee.bytes", validator.MinInt(f.Fee.Standard.MiningFee.Bytes, 0)).
		Validate("fee.standard.relayFee.satsohis", validator.MinInt(f.Fee.Standard.RelayFee.Satoshis, 0)).
		Validate("fee.standard.relayFee.bytes", validator.MinInt(f.Fee.Standard.RelayFee.Bytes, 0)).
		Validate("fee.data.miningFee.satsohis", validator.MinInt(f.Fee.Data.MiningFee.Satoshis, 0)).
		Validate("fee.data.miningFee.bytes", validator.MinInt(f.Fee.Data.MiningFee.Bytes, 0)).
		Validate("fee.data.relayFee.satsohis", validator.MinInt(f.Fee.Data.RelayFee.Satoshis, 0)).
		Validate("fee.data.relayFee.bytes", validator.MinInt(f.Fee.Data.RelayFee.Bytes, 0)).
		Validate("tx", validator.NotEmpty(f.TxHex), validator.IsHex(f.TxHex), func() error {
			tx, err := bt.NewTxFromString(f.TxHex)
			if err != nil {
				return err
			}
			if len(tx.Inputs) > 0 {
				return errors.New("tx cannot already be funded")
			}
			return nil
		})

	return v.Err()
}

// SignTxResponse a signed tx response.
type SignTxResponse struct {
	SignedTx string `json:"signedTx"`
}

package payd

import (
	"context"
	"time"

	"github.com/libsv/go-bt/v2"
)

type PayRequest struct {
	PayToURL string `json:"payToURL"`
}

type P4Output struct {
	Amount      uint64 `json:"amount"`
	Script      string `json:"script"`
	Description string `json:"description"`
}
type PaymentRequestResponse struct {
	Network             string     `json:"network"`
	Outputs             []P4Output `json:"outputs"`
	CreationTimestamp   time.Time  `json:"creationTimestamp"`
	ExpirationTimestamp time.Time  `json:"expirationTimestamp"`
	PaymentURL          string     `json:"paymentURL"`
	Memo                string     `json:"memo"`
	MerchantData        struct {
		Avatar           string            `json:"avatar"`
		Name             string            `json:"name"`
		Email            string            `json:"email"`
		Address          string            `json:"address"`
		PaymentReference string            `json:"paymentReference"`
		ExtendedData     map[string]string `json:"extendedData"`
	} `json:"merchantData"`
	Fee *bt.FeeQuote `json:"fee"`
}

type PayService interface {
	Pay(ctx context.Context, req PayRequest) error
}

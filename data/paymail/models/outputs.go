package models

import (
	gopaymail "github.com/tonicpow/go-paymail"

	gopayd "github.com/libsv/payd"
)

func OutputsToPayd(oo []*gopaymail.PaymentOutput) []*gopayd.Output {
	out := make([]*gopayd.Output, 0)
	for _, o := range oo {
		out = append(out, &gopayd.Output{
			Amount:      o.Satoshis,
			Script:      o.Script,
			Description: "",
		})
	}
	return out
}

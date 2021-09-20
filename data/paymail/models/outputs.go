package models

import (
	"github.com/libsv/go-bt/v2"
	gopaymail "github.com/tonicpow/go-paymail"
)

// OutputsToPayd will convert the paymail data type to a payd output.
func OutputsToPayd(oo []*gopaymail.PaymentOutput) []*bt.Output {
	out := make([]*bt.Output, 0)
	for _, o := range oo {
		out = append(out, &bt.Output{
			Satoshis: o.Satoshis,
			//LockingScript:    bscript.  o.Script,
		})
	}
	return out
}

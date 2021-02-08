package gopayd

import (
	"testing"

	"github.com/matryer/is"
)

func TestCreatePayment_Validate(t *testing.T) {
	is := is.New(t)
	tests := map[string]struct {
		req CreatePayment
		exp string
	}{
		"valid request should return no errors": {
			req: CreatePayment{
				Transaction:  "74657374696e672074657374696e67",
				MerchantData: func() *string { s := "some data"; return &s }(),
				RefundTo:     nil,
				Memo:         "test this please",
			},
			exp: "no validation errors",
		}, "transaction with invalid prefix should error": {
			req: CreatePayment{
				Transaction:  "0x74657374696e672074657374696e67",
				MerchantData: func() *string { s := "some data"; return &s }(),
				RefundTo:     nil,
				Memo:         "test this please",
			},
			exp: "[transaction: value provided does not have a valid prefix, value supplied is not valid hex]",
		}, "transaction with invalid hex should error": {
			req: CreatePayment{
				Transaction:  "74657374696e672074657374696ezz67",
				MerchantData: nil,
				RefundTo:     nil,
				Memo:         "test this please",
			},
			exp: "[transaction: value supplied is not valid hex]",
		}, "merchant data too long should error": {
			req: CreatePayment{
				Transaction: "74657374696e672074657374696e67",
				MerchantData: func() *string {
					bb := make([]byte, 0)
					// generate string 1 more byte than 10000
					for i := 0; i <= 10000; i++ {
						bb = append(bb, 42)
					}
					o := string(bb)
					return &o
				}(),
				RefundTo: nil,
				Memo:     "test this please",
			},
			exp: "[merchantData: value must be between 0 and 10000 characters]",
		}, "refundTo too long should error": {
			req: CreatePayment{
				Transaction:  "74657374696e672074657374696e67",
				MerchantData: nil,
				RefundTo: func() *string {
					bb := make([]byte, 0)
					// generate string 1 more byte than 10000
					for i := 0; i <= 100; i++ {
						bb = append(bb, 42)
					}
					o := string(bb)
					return &o
				}(),
				Memo: "test this please",
			},
			exp: "[refundTo: value must be between 0 and 100 characters]",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			is.NewRelaxed(t)
			is.Equal(test.req.Validate() != nil, true)
			is.Equal(test.exp, test.req.Validate().Error())
		})
	}
}

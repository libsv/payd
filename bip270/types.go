package bip270

// Output message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#output
type Output struct {
	Amount      uint64 `json:"amount"`                // satoshis
	Script      string `json:"script"`                // locking script
	Description string `json:"description,omitempty"` // must not have JSON string length of greater than 100.
}

// PaymentRequest message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#paymentrequest
type PaymentRequest struct {
	Network             string        `json:"network"` // always set to "bitcoin" (but seems to be set to 'bitcoin-sv' outside bip270 spec, see https://handcash.github.io/handcash-merchant-integration/#/merchant-payments)
	Outputs             []*Output     `json:"outputs"` // an array of outputs. required, but can have zero elements.
	CreationTimestamp   int64         `json:"creationTimestamp"`
	ExpirationTimestamp int64         `json:"expirationTimestamp,omitempty"`
	PaymentURL          string        `json:"paymentUrl"`
	Memo                string        `json:"memo,omitempty"`
	MerchantData        *MerchantData `json:"merchantData,omitempty"`
}

// MerchantData to be displayed to the user.
type MerchantData struct {
	AvatarURL    string `json:"avatarUrl,omitempty"`
	MerchantName string `json:"merchantName,omitempty"`
}

// Payment message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#payment
type Payment struct {
	Transaction  string `json:"transaction"`
	MerchantData string `json:"merchantData,omitempty"`
	RefundTo     string `json:"refundTo,omitempty"`
	Memo         string `json:"memo,omitempty"`
}

// PaymentACK message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#paymentack
type PaymentACK struct {
	Payment *Payment `json:"payment"`
	Memo    string   `json:"memo,omitempty"`
	// A number indicating why the transaction was not accepted. 0 or undefined indicates no error.
	// A 1 or any other positive integer indicates an error. The errors are left undefined for now;
	// it is recommended only to use “1” and to fill the memo with a textual explanation about why
	// the transaction was not accepted until further numbers are defined and standardized.
	Error int `json:"error,omitempty"`
	// TODO: check anypay https://docs.anypayinc.com/pay-protocol/overview
	// and consider deleting success field because seems they don't actually
	// return it, tried sending a rubbish tx and didn't get back success: "false"
	// - only got back proper bip270 paymentack response (like above)
	Success string `json:"success,omitempty"` // 'true' or 'false' string
}

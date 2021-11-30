package p4

import (
	"context"
)

// Merchant to be displayed to the user.
type Merchant struct {
	// AvatarURL displays a canonical url to a merchants avatar.
	AvatarURL string `json:"avatar" example:"http://url.com"`
	// Name is a human readable string identifying the merchant.
	Name string `json:"name" example:"merchant 1"`
	// Email can be sued to contact the merchant about this transaction.
	Email string `json:"email" example:"merchant@m.com"`
	// Address is the merchants store / head office address.
	Address string `json:"address" example:"1 the street, the town, B1 1AA"`
	// ExtendedData can be supplied if the merchant wishes to send some arbitrary data back to the wallet.
	ExtendedData map[string]interface{} `json:"extendedData,omitempty"`
}

// MerchantReader is used to read merchant data from a data store or service.
type MerchantReader interface {
	// Owner will return MerchantData from a data store, owner being the person who owns the wallet.
	Owner(ctx context.Context) (*Merchant, error)
}

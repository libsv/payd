package wallet

import (
	"context"
	"time"

	"github.com/bitcoinsv/bsvutil/hdkeychain"
)

// PrivateKey describes a named private key.
type PrivateKey struct {
	// Name of the private key.
	Name string `db:"name"`
	// Xpriv is the private key.
	Xpriv string `db:"xpriv"`
	// CreatedAt is the date/time when the key was stored.
	CreatedAt time.Time `db:"createdAt"`
}

// KeyArgs defines all arguments required to get a key.
type KeyArgs struct {
	// Name is the name of the key to return.
	Name string
}

// PrivateKeyService can be implemented to get and create PrivateKeys.
type PrivateKeyService interface {
	// Create will create a new private key if it doesn't exist already.
	Create(ctx context.Context, keyName string) error
	// PrivateKey will return a private key.
	PrivateKey(ctx context.Context, keyName string) (*hdkeychain.ExtendedKey, error)
	// DeriveChildFromKey will create a private key derived from a parent key at the given derivationPath.
	DeriveChildFromKey(startingKey *hdkeychain.ExtendedKey, derivationPath string) (*hdkeychain.ExtendedKey, error)
	// PubFromXPrv will generate a public key from an extended private key.
	PubFromXPrv(xprv *hdkeychain.ExtendedKey) ([]byte, error)
}

// KeyStorer describes a data store that can be implemented to get and store private keys.
type KeyStorer interface {
	// PrivateKey can be used to return an existing private key.
	PrivateKey(ctx context.Context, args KeyArgs) (*PrivateKey, error)
	// Create will add a new private key to the data store.
	Create(ctx context.Context, req PrivateKey) (*PrivateKey, error)
}

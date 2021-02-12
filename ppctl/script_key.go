package ppctl

import (
	"context"
)

// ScriptKey contains data required to sign a
// received transaction output.
// This is an internal data structure.
type ScriptKey struct {
	ID             int    `json:"-" db:"id"`
	LockingScript  string `json:"-" db:"id"`
	KeyName        string `json:"-" db:"id"`
	DerivationPath string `json:"-" db:"id"`
}

// CreateScriptKey can be used to create a new ScriptKey.
type CreateScriptKey struct {
	LockingScript  string `db:"lockingscript"`
	KeyName        string `db:"keyname"`
	DerivationPath string `db:"derivationpath"`
}

// ScriptKeyArgs contain arguments used to get a script key
// to validate a utxo.
type ScriptKeyArgs struct {
	LockingScript string `db:"lockingscript"`
}

// ScriptKeyStorer can be implemented to store and return ScriptKeys.
type ScriptKeyStorer interface {
	Create(ctx context.Context, req []CreateScriptKey) error
	ScriptKey(ctx context.Context, args ScriptKeyArgs) (*ScriptKey, error)
}

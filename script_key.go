package gopayd

import (
	"context"

	"gopkg.in/guregu/null.v3"
)

// ScriptKey contains data required to sign a
// received transaction output.
// This is an internal data structure.
type ScriptKey struct {
	ID             int         `json:"-" db:"id"`
	LockingScript  string      `json:"-" db:"lockingscript"`
	KeyName        null.String `json:"-" db:"keyname"`
	DerivationPath null.String `json:"-" db:"path"`
}

// CreateScriptKey can be used to create a new ScriptKey.
type CreateScriptKey struct {
	LockingScript string      `db:"lockingscript"`
	KeyName       null.String `db:"keyname"`
	DerivationID  null.Int    `db:"derivationID"`
}

// ScriptKeyArgs contain arguments used to get a script key
// to validate a utxo.
type ScriptKeyArgs struct {
	LockingScript string `db:"lockingscript"`
}

// ScriptKeyReaderWriter can be implemented to store and return ScriptKeys.
type ScriptKeyReaderWriter interface {
	ScriptKeyWriter
	ScriptKeyReader
}

// ScriptKeyWriter can be implemented to store and return ScriptKeys.
type ScriptKeyWriter interface {
	CreateScriptKeys(ctx context.Context, req []CreateScriptKey) error
}

// ScriptKeyReader reads script key info from a data store.
type ScriptKeyReader interface {
	// ScriptKey will return a scriptKey matching the args field.
	ScriptKey(ctx context.Context, args ScriptKeyArgs) (*ScriptKey, error)
}

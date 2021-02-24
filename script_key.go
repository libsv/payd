package gopayd

import (
	"context"
)

// ScriptKey contains data required to sign a
// received transaction output.
// This is an internal data structure.
type ScriptKey struct {
	ID             int    `json:"-" db:"id"`
	LockingScript  string `json:"-" db:"lockingscript"`
	KeyName        string `json:"-" db:"keyname"`
	DerivationPath string `json:"-" db:"path"`
}

// CreateScriptKey can be used to create a new ScriptKey.
type CreateScriptKey struct {
	LockingScript string `db:"lockingscript"`
	KeyName       string `db:"keyname"`
	DerivationID  int    `db:"derivationID"`
}

// ScriptKeyArgs contain arguments used to get a script key
// to validate a utxo.
type ScriptKeyArgs struct {
	LockingScript string `db:"lockingscript"`
}

// ScriptKeyStorer can be implemented to store and return ScriptKeys.
type ScriptKeyReaderWriter interface {
	ScriptKeyWriter
	ScriptKeyReader
}

// ScriptKeyStorer can be implemented to store and return ScriptKeys.
type ScriptKeyWriter interface {
	CreateScriptKeys(ctx context.Context, req []CreateScriptKey) error
}

type ScriptKeyReader interface {
	// ScriptKey will return a scriptKey matching the args field.
	ScriptKey(ctx context.Context, args ScriptKeyArgs) (*ScriptKey, error)
}

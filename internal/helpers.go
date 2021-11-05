package internal

import (
	"github.com/libsv/go-bt/v2/bscript"
)

// StringToScript will convert a string to a bscript ignoring errors, mostly used for test funcs.
func StringToScript(s string) *bscript.Script {
	sc, _ := bscript.NewFromHexString(s)
	return sc
}

package prnt

import (
	"encoding/json"
	"fmt"
)

type jsonPrint struct{}

func NewJSONPrinter() Printer {
	return &jsonPrint{}
}

func (j *jsonPrint) Print(v interface{}) error {
	bb, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(bb))
	return nil
}

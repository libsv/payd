package output

import "fmt"

// PrintFunc defines a printer func.
type PrintFunc func(v interface{}) error

// NewPrinter returns a PrinterFunc.
func NewPrinter(f Format) PrintFunc {
	var fn PrintFunc
	switch f {
	case FormatTable:
		fn = tablePrint
	case FormatJSON:
		fn = jsonPrint
	case FormatYaml:
		fn = yamlPrint
	default:
		fn = tablePrint
	}

	return func(v interface{}) error {
		if s, ok := v.(string); ok {
			fmt.Println(s)
			return nil
		}

		return fn(v)
	}
}

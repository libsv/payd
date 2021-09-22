package output

import "fmt"

var strats map[Format]PrintFunc = map[Format]PrintFunc{
	FormatTable: tablePrint,
	FormatJSON:  jsonPrint,
	FormatYaml:  yamlPrint,
}

// PrintFunc defines a printer func.
type PrintFunc func(v interface{}) error

// NewPrinter returns a PrinterFunc.
func NewPrinter(f Format) PrintFunc {
	return func(v interface{}) error {
		if s, ok := v.(string); ok {
			fmt.Println(s)
			return nil
		}

		return strats[f](v)
	}
}

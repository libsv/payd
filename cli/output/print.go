package output

import "fmt"

type PrintFunc func(v interface{}) error

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
		if err, ok := v.(error); ok {
			fmt.Println(err.Error())
			return nil
		}

		return fn(v)
	}
}

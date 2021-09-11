package prnt

type Format string

const (
	Table Format = "table"
	JSON  Format = "json"
)

func NewPrinter(f Format) Printer {
	switch f {
	case Table:
		return NewTablePrinter()
	case JSON:
		return NewJSONPrinter()
	}

	return NewTablePrinter()
}

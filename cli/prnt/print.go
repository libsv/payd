package prnt

type Printer interface {
	Print(v interface{}) error
}

type Tableable interface {
	Columns() []string
	Rows() [][]string
}

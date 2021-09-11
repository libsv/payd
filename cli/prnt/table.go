package prnt

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
)

type table struct {
}

func NewTablePrinter() Printer {
	return &table{}
}

func (t *table) Print(v interface{}) error {
	tab := reflect.ValueOf(v).Interface().(Tableable)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 4, '\t', 0)
	fmt.Fprintln(w, strings.Join(tab.Columns(), "\t"))
	for _, row := range tab.Rows() {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	fmt.Fprintln(w)
	w.Flush()
	return w.Flush()
}

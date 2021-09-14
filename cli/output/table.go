package output

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
)

type Tableable interface {
	Columns() []string
	Rows() [][]string
}

func tablePrint(v interface{}) error {
	tab, ok := reflect.ValueOf(v).Interface().(Tableable)
	if !ok {
		fmt.Fprintln(os.Stderr, "table not implemented for resource, falling back to json")
		return jsonPrint(v)
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	fmt.Fprintln(w, strings.Join(tab.Columns(), "\t"))
	for _, row := range tab.Rows() {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	fmt.Fprintln(w)
	return w.Flush()
}

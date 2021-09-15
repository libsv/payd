package output

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Wrapped interfaces a wrapped data object.
type Wrapped interface {
	Unwrap() interface{}
}

func yamlPrint(v interface{}) error {
	wrapped, ok := reflect.ValueOf(v).Interface().(Wrapped)
	if ok {
		v = wrapped.Unwrap()
	}

	bb, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	fmt.Println(string(bb))
	return nil
}

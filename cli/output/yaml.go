package output

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Wrap interfaces a wrap data object.
type Wrap interface {
	Unwrap() interface{}
}

func yamlPrint(v interface{}) error {
	unwrapped, ok := reflect.ValueOf(v).Interface().(Wrap)
	if ok {
		v = unwrapped.Unwrap()
	}

	bb, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	fmt.Println(string(bb))
	return nil
}

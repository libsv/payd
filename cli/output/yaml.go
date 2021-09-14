package output

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func yamlPrint(v interface{}) error {
	bb, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	fmt.Println(string(bb))
	return nil
}

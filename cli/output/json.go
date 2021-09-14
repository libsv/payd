package output

import (
	"encoding/json"
	"fmt"
)

func jsonPrint(v interface{}) error {
	bb, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(bb))
	return nil
}

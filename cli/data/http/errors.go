package http

import (
	"encoding/json"
	"net/http"

	"github.com/libsv/payd/cli/models"
)

func checkError(resp *http.Response, statusCode int) error {
	if resp.StatusCode == statusCode {
		return nil
	}

	var errResp models.ServerError
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return err
	}

	return errResp
}

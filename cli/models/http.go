package models

import "net/http"

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

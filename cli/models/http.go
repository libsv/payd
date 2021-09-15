package models

import "net/http"

// HTTPClient interfaces the *http.Client.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

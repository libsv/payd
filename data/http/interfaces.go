package http

import "net/http"

// Client interfaces the Do(*http.Request) function to allow for easy mocking.
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

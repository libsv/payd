package models

import (
	"fmt"
)

type ServerError struct {
	ID      string `json:"id"`
	Code    string `json:"code"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

func (s ServerError) Error() string {
	return fmt.Sprintf("%s %s: %s, %s", s.Code, s.ID, s.Title, s.Message)
}

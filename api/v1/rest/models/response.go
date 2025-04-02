package models

type ApiResponse[K string, T any] struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    map[string]T `json:"data,omitempty"`
}

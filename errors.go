package stupid

import (
	"net/http"
)

type StupidError struct {
	code int
}

func NewStupidError(statusCode int) StupidError {
	return StupidError{code: statusCode}
}

func (s StupidError) Error() string {
	switch s.code {
	case http.StatusBadRequest:
		return "bad request"
	case http.StatusUnauthorized:
		return "unauthorized"
	default:
		return "unknown"
	}
}

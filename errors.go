package stupid

import (
	"net/http"
)

type StupidError struct {
	code int
}

// Create a new StupidError wrapping the given http status code.
func NewStupidError(statusCode int) StupidError {
	return StupidError{code: statusCode}
}

func (s StupidError) Error() string {
	switch s.code {
	case http.StatusBadRequest:
		return "bad request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusNoContent:
		return "no content"
	default:
		return "unknown"
	}
}

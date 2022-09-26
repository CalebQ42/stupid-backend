package stupid

import "errors"

var (
	ErrBadRequest   = errors.New("bad request")  //Sets return header to http.ErrBadRequest
	ErrUnauthorized = errors.New("unauthorized") //Sets return header to http.ErrUnauthorized
)

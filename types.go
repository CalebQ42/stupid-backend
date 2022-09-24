package stupid

import (
	"io"
	"net/url"
)

type DataRequestHandler interface {
	Request(query url.Values, body io.ReadCloser) ([]byte, error)
	AuthenticatedRequest(uuid string, query url.Values, body io.ReadCloser) ([]byte, error)
}

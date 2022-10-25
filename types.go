package stupid

import (
	"io"
	"net/url"

	"go.mongodb.org/mongo-driver/mongo"
)

// A Request that isn't handled directly by stupid.Backend
type Request struct {
	ReqBody     io.ReadCloser   // The request's body. Might be empty or nil. Will be closed at the end of the function, so it might be necessary to copy it's contents if concerency is needed.
	Query       url.Values      // The request's URL Query.
	KeyFeatures map[string]bool // The features of the API Key that made the request.
	User        *RequestUser    // The authenticated user if the request was authenticated with a token. Will be nil if no token was provided.
	Method      string          // Request's method (POST, GET, etc)
	Path        string          // Request's URL path.
}

// A basic app for the backend. ID() is expected to be unique for each application that the backend uses.
type App interface {
	ID() string
	Logs() *mongo.Collection
	Crashes() *mongo.Collection
	Initialize() error
}

// An app that additionally can handle data requests (urls that start with /data).
// DataRequest is only called if the API Key used in the request is valid and if the token provided is valid (if one is provided).
type DataApp interface {
	App
	DataRequest(req *Request) (body []byte, err error) // If err is set to a StupidError, it will set the header to the coresponding code. Otherwise if err is non null, it will be logged and internal error will be sent.
}

// An optional extension on stupid.Backend to handle new requests.
type BackendExtension interface {
	HandleRequest(request *Request) (body []byte, err error) // If err is set to a StupidError, it will set the header to the coresponding code. Otherwise if err is non null, it will be logged and internal error will be sent.
}

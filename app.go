package stupid

import (
	"io"
	"net/url"

	"go.mongodb.org/mongo-driver/mongo"
)

// A basic app for the backend. ID() is expected to be unique for each application that the backend uses.
type App interface {
	ID() string
	Crashes() *mongo.Collection
	Logs() *mongo.Collection
	Initialize() error
}

// A request for data. Used for both authenticated and unauthenticated requests.
type DataRequest struct {
	ReqBody     io.ReadCloser   // The request's body. Might be empty or nil.
	Query       url.Values      // The request's URL Query.
	KeyFeatures map[string]bool // The features of the API Key that made the request.
	UserID      string          // The UUID of the user if the request was authenticated with a token. Will be an empty if no token was provided.
}

// An app that additionally can handle data requests.
// DataRequest is only called if the API Key used in the request is valid and if the token provided is valid (if one is provided).
type DataApp interface {
	App
	DataRequest(req DataRequest) ([]byte, error)
}

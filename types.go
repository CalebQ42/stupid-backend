package stupid

import (
	"crypto/ed25519"
	"io"
	"net/url"

	"go.mongodb.org/mongo-driver/mongo"
)

// A Request that isn't handled directly by stupid.Backend
type Request struct {
	ReqBody     io.ReadCloser   // The request's body. Might be empty or nil.
	Query       url.Values      // The request's URL Query.
	KeyFeatures map[string]bool // The features of the API Key that made the request.
	Path        string          // Request's URL path.
	UserID      string          // The UUID of the user if the request was authenticated with a token. Will be an empty if no token was provided.
}

// A basic app for the backend. ID() is expected to be unique for each application that the backend uses.
type App interface {
	ID() string
	Logs() *mongo.Collection
	Crashes() *mongo.Collection
	Initialize() error
}

// An app that additionally can handle data requests.
// DataRequest is only called if the API Key used in the request is valid and if the token provided is valid (if one is provided).
type DataApp interface {
	App
	DataRequest(req Request) (body []byte, err error) // If err is set to a stupid error type, it will set the header to the coresponding code. Otherwise if err is non null, it will be logged and internal error will be sent.
}

// A DataApp that allows for authenticated requests. Given keys are used to authenticate JWT tokens when making data requests.
type AuthenticatedDataApp interface {
	DataApp
	PublicJWTKey() ed25519.PublicKey
	PrivateJWTKey() ed25519.PrivateKey
}

// An optional extension on stupid.Backend to handle new requests.
type BackendExtension interface {
	HandleRequest(request Request) (body []byte, err error) // If err is set to a stupid error type, it will set the header to the coresponding code. Otherwise if err is non null, it will be logged and internal error will be sent.
}

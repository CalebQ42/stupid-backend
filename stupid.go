package stupid

import (
	"crypto/ed25519"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/CalebQ42/stupid-backend/pkg/db"
)

// An instance of the stupid backend. Implements http.Handler
type Stupid struct {
	keys  db.Table
	users db.UserTable
	// Get a db.App for the given appId.
	Apps func(appID string) App

	createUserMutex *sync.Mutex
	headerValues    map[string]string
	userPriv        ed25519.PrivateKey
	userPub         ed25519.PublicKey
}

// Creates a new *Stupid.
func NewStupidBackend(keyTable db.Table, apps func(appID string) App) *Stupid {
	return &Stupid{
		keys:            keyTable,
		createUserMutex: &sync.Mutex{},
		Apps:            apps,
	}
}

// Adds header values to add to every response (such as Access-Control-Allow-Origin).
func (s *Stupid) SetHeaderValues(values map[string]string) {
	s.headerValues = values
}

// Enables user authentication and creation.
func (s *Stupid) EnableUserAuth(userTable db.UserTable, pubKey ed25519.PublicKey, privKey ed25519.PrivateKey) {
	s.users = userTable
	s.userPub = pubKey
	s.userPriv = privKey
}

// Satisfies http.Handler
func (s *Stupid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &Request{
		Body:   r.Body,
		Query:  r.URL.Query(),
		Path:   strings.Split(strings.TrimPrefix(path.Clean(r.URL.Path), "/"), "/"),
		Method: r.Method,
		Resp:   w,
	}
	if s.headerValues != nil {
		for k, v := range s.headerValues {
			w.Header().Set(k, v)
		}
	}
	if len(req.Path) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !req.validKey(s.keys) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	err := s.handleToken(req)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	app := s.Apps(req.ApiKey.AppID)
	if app == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch req.Path[0] {
	case "key":
		if req.ApiKey.Permissions["key"] {
			req.handleKeyReq()
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	case "log":
		if req.ApiKey.Permissions["log"] {
			s.logReq(req, app.Logs())
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	case "crash":
		if req.ApiKey.Permissions["crash"] {
			s.crashReport(req, app.Crashes())
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	case "createUser":
		if req.ApiKey.Permissions["auth"] {
			s.createUser(req)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	case "auth":
		if req.ApiKey.Permissions["auth"] {
			s.authUser(req)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	default:
		if !app.Extension(req) {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

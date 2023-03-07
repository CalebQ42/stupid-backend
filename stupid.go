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
	AppTables func(appID string) db.App
	// Handle a request beyond the default abilities.
	Extension       func(*Request)
	createUserMutex *sync.Mutex
	userPriv        ed25519.PrivateKey
	userPub         ed25519.PublicKey
}

func NewStupidBackend(keyTable db.Table) *Stupid {
	return &Stupid{
		keys:            keyTable,
		createUserMutex: &sync.Mutex{},
	}
}

func (s *Stupid) EnableUserAuth(userTable db.UserTable, pubKey ed25519.PublicKey, privKey ed25519.PrivateKey) {
	s.users = userTable
	s.userPub = pubKey
	s.userPriv = privKey
}

// Sets *Stupid.AppTables to use this map, overriding if it's already been set.
func (s *Stupid) SetApps(apps map[string]db.App) {
	s.AppTables = func(id string) db.App {
		return apps[id]
	}
}

func (s *Stupid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &Request{
		Body:   r.Body,
		Query:  r.URL.Query(),
		Path:   strings.Split(strings.TrimPrefix(path.Clean(r.URL.Path), "/"), "/"),
		Method: r.Method,
		Resp:   w,
	}
	if len(req.Path) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !req.validKey(s.keys) {
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
		if s.AppTables != nil && req.ApiKey.Permissions["log"] {
			s.logReq(req, s.AppTables(req.ApiKey.AppID).Logs)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	case "crash":
		if s.AppTables != nil && req.ApiKey.Permissions["crash"] {
			s.crashReport(req, s.AppTables(req.ApiKey.AppID).Crashes)
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
			//TODO: autnenticate user
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

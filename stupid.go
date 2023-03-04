package stupid

import (
	"net/http"
	"path"
	"strings"

	"github.com/CalebQ42/stupid-backend/pkg/db"
)

// An instance of the stupid backend. Implements http.Handler
type Stupid struct {
	Keys  db.Table
	Users db.Table
	// Get a db.App for the given appId.
	AppTables func(appID string) db.App

	Extension func(*Request)
}

func NewStupidBackend(keyTable db.Table) *Stupid {
	return &Stupid{
		Keys: keyTable,
	}
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
	if !req.validKey(s.Keys) {
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
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

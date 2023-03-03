package stupid

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/CalebQ42/stupid-backend/pkg/crash"
	"github.com/CalebQ42/stupid-backend/pkg/db"
)

// An instance of the stupid backend. Implements http.Handler
type Stupid struct {
	Keys  db.Table
	Users db.Table
	// Get a db.App for the given appId.
	AppTables func(appID string) db.App
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
	req := &stupidRequest{
		r:      r.Body,
		query:  r.URL.Query(),
		path:   strings.Split(strings.TrimPrefix(path.Clean(r.URL.Path), "/"), "/"),
		method: r.Method,
		w:      w,
	}
	if len(req.path) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !req.validKey(s.Keys) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch req.path[0] {
	case "key":
		if req.apiKey.Permissions["key"] {
			req.handleKeyReq()
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	case "log":
		if s.AppTables != nil && req.apiKey.Permissions["log"] {
			s.logReq(req, s.AppTables(req.apiKey.AppID).Logs)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	case "crash":
		if s.AppTables != nil && req.apiKey.Permissions["crash"] {
			s.crashReport(req, s.AppTables(req.apiKey.AppID).Crashes)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *Stupid) logReq(req *stupidRequest, logs db.Table) {
	if req.method != http.MethodPost {
		req.w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, ok := req.query["id"]
	if !ok || len(id) != 1 {
		req.w.WriteHeader(http.StatusBadRequest)
		return
	}
	plat, ok := req.query["platform"]
	if !ok || len(plat) != 1 {
		req.w.WriteHeader(http.StatusBadRequest)
		return
	}
	ok, err := logs.Has(id[0])
	if err != nil {
		req.w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error while checking if log id is already present: %s", err)
		return
	}
	usr := logUser{
		ID:       id[0],
		Platform: plat[0],
		LastCon:  time.Now(),
	}
	if ok {
		err = logs.Update(id[0], usr)
		if err != nil {
			req.w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error while updating log: %s", err)
			return
		}
	} else {
		_, err = logs.Add(usr)
		if err != nil {
			req.w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error while adding log: %s", err)
			return
		}
	}
	req.w.WriteHeader(http.StatusCreated)
}

func (s *Stupid) crashReport(req *stupidRequest, table db.CrashTable) {
	if req.method != http.MethodPost {
		req.w.WriteHeader(http.StatusBadRequest)
		return
	}
	dec := json.NewDecoder(req.r)
	var c crash.Individual
	err := dec.Decode(&c)
	req.r.Close()
	if err != nil {
		req.w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = table.AddCrash(c)
	if err != nil {
		req.w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.w.WriteHeader(http.StatusCreated)
}

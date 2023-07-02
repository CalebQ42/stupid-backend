package stupid

import (
	"crypto/ed25519"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/CalebQ42/stupid-backend/pkg/db"
)

// An instance of the stupid backend. Implements http.Handler
type Stupid struct {
	keys            db.Table
	users           db.UserTable
	Apps            map[string]App
	createUserMutex *sync.Mutex
	headerValues    map[string]string
	userPriv        ed25519.PrivateKey
	userPub         ed25519.PublicKey
	cors            bool
}

// Creates a new *Stupid.
func NewStupidBackend(keyTable db.Table, apps map[string]App, allowCors bool) *Stupid {
	out := &Stupid{
		keys:            keyTable,
		createUserMutex: &sync.Mutex{},
		Apps:            apps,
		cors:            allowCors,
	}
	go out.cleanupLoop()
	return out
}

func (s *Stupid) cleanupLoop() {
	for range time.Tick(24 * time.Hour) {
		log.Println("Cleaning up old logs")
		cleanTmp := time.Now().Add(-24 * 30 * time.Hour)
		cleanVal := cleanTmp.Year()*10000 + int(cleanTmp.Month())*100 + cleanTmp.Day()
		var ids []string
		var err error
		for appName := range s.Apps {
			ids, err = s.Apps[appName].Logs().LogsOlderThen(cleanVal)
			if err != nil {
				log.Println("Error when cleaning up old logs for "+appName+":", err)
				continue
			}
			for i := range ids {
				err = s.Apps[appName].Logs().Delete(ids[i])
				if err != nil {
					log.Println("Error when deleting old logs for "+appName+":", err)
				}
			}
		}
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
	if s.cors && r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
		return
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
	app := s.Apps[req.ApiKey.AppID]
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
	case "count":
		if req.ApiKey.Permissions["count"] {
			s.count(req, app.Logs())
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

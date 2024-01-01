package stupid

import (
	"crypto/ed25519"
	_ "embed"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/CalebQ42/stupid-backend/v2/db"
)

//go:embed embed/robots.txt
var robotsTxt []byte

// An instance of the stupid backend. Implements http.Handler
type Stupid struct {
	keys            db.Table
	users           db.UserTable
	Apps            map[string]any
	createUserMutex *sync.Mutex
	headerValues    map[string]string
	cors            string
	userPriv        ed25519.PrivateKey
	userPub         ed25519.PublicKey
}

// Creates a new *Stupid. If corsAddress is empty, CORS is not allowed.
func NewStupidBackend(keyTable db.Table, apps map[string]any, corsAddress string) *Stupid {
	for appName, a := range apps {
		if _, ok := a.(KeyedApp); ok {
			continue
		} else if ua, ok := a.(UnKeyedApp); ok {
			if alt, ok := ua.(UnKeyedWithAlternateNameApp); ok {
				_, exists := apps[alt.AlternateName()]
				if exists {
					log.Fatalln("Alternate name for", appName, "already exists")
				} else {
					apps[alt.AlternateName()] = ua
				}
			}
			continue
		}
		log.Fatalln("App", a, "does not implement KeyedApp or UnKeyedApp")
	}
	out := &Stupid{
		keys:            keyTable,
		createUserMutex: &sync.Mutex{},
		Apps:            apps,
		cors:            corsAddress,
	}
	go out.cleanupLoop()
	return out
}

func (s *Stupid) cleanupLoop() {
	s.cleanup()
	for range time.Tick(24 * time.Hour) {
		s.cleanup()
	}
}

func (s *Stupid) cleanup() {
	log.Println("Cleaning up old logs")
	cleanTmp := time.Now().Add(-30 * 24 * time.Hour) // 30 days prior
	cleanVal := cleanTmp.Year()*10000 + int(cleanTmp.Month())*100 + cleanTmp.Day()
	var ids []string
	var err error
	for appName, a := range s.Apps {
		app, ok := a.(KeyedApp)
		if !ok {
			continue
		}
		ids, err = app.Logs().LogsOlderThen(cleanVal)
		if err != nil {
			log.Println("Error when cleaning up old logs for "+appName+":", err)
			continue
		}
		for i := range ids {
			err = app.Logs().Delete(ids[i])
			if err != nil {
				log.Println("Error when deleting old logs for "+appName+":", err)
			}
		}
		log.Println("Cleaned up", len(ids), "logs for", appName)
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
	if len(req.Path) == 1 && req.Path[0] == "robots.txt" {
		w.Write(robotsTxt)
		return
	}
	if s.headerValues != nil {
		for k, v := range s.headerValues {
			w.Header().Set(k, v)
		}
	}
	if s.cors != "" {
		w.Header().Set("Access-Control-Allow-Origin", s.cors)
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
			return
		}
	}
	if len(req.Path) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !req.validKey(s.keys) {
		_, ok := req.Query["key"]
		if !ok {
			if s.handlePossibleUnKeyedApp(req) {
				return
			}
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	err := s.handleToken(req)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	a := s.Apps[req.ApiKey.AppID]
	if a == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	app, ok := a.(KeyedApp)
	if !ok {
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
		ext, ok := app.(ExtendedApp)
		if !ok || !ext.Extension(req) {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func (s *Stupid) handlePossibleUnKeyedApp(req *Request) bool {
	app, ok := s.Apps[req.Path[0]].(UnKeyedApp)
	if !ok {
		return false
	}
	if !app.HandleReqest(req) {
		req.Resp.WriteHeader(http.StatusBadRequest)
	}
	return true
}

package stupid

import (
	"net/http"
	"path"
	"strings"

	"github.com/CalebQ42/stupid-backend/pkg/db"
)

// An instance of the stupid backend. Implements http.Handler
type Stupid struct {
	Keys db.DBTable
	Logs db.DBTable
}

func (s *Stupid) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &stupidRequest{
		r:     r.Body,
		query: r.URL.Query(),
		path:  strings.Split(strings.TrimPrefix(path.Clean(r.URL.Path), "/"), "/"),
		w:     w,
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
		req.handleKeyReq()
	case "log":
		s.log(req)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *Stupid) log(req *stupidRequest) {
	
}

package stupid

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/CalebQ42/stupid-backend/pkg/db"
)

type Request struct {
	Body   io.ReadCloser
	Resp   http.ResponseWriter
	Query  map[string][]string
	User   *AuthdUser
	Method string
	ApiKey apiKey
	Path   []string
}

// Both validates that the api key is present and is valid and
// populates the apiKey value of stupidRequest (if it is valid).
func (s *Request) validKey(keyTable db.Table) bool {
	var key string
	if s.Path[0] == "key" {
		if len(s.Path) == 1 {
			return false
		}
		key = s.Path[1]
	} else {
		k, ok := s.Query["key"]
		if !ok || len(k) != 1 {
			return false
		}
		key = k[0]
	}
	err := keyTable.Get(key, &s.ApiKey)
	if err != nil {
		return false
	}
	if s.ApiKey.Death != -1 {
		deth := time.Unix(s.ApiKey.Death, 0)
		if time.Now().After(deth) {
			return false
		}
	}
	return true
}

func (s *Request) handleKeyReq() {
	if s.Method != http.MethodGet {
		s.Resp.WriteHeader(http.StatusBadRequest)
		return
	}
	out, err := json.MarshalIndent(s.ApiKey, "", "\t")
	if err != nil {
		log.Printf("error while marshalling key: %s", err)
		s.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.Resp.Header().Add("content-type", "application/json")
	_, err = s.Resp.Write(out)
	if err != nil {
		log.Printf("error while writing key: %s", err)
		s.Resp.WriteHeader(http.StatusInternalServerError)
	}
}

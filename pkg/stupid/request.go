package stupid

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/CalebQ42/stupid-backend/pkg/db"
)

type stupidRequest struct {
	r         io.ReadCloser
	w         http.ResponseWriter
	query     map[string][]string
	authdUser string
	apiKey    apiKey
	path      []string
}

// Both validates that the api key is present and is valid and
// populates the apiKey value of stupidRequest (if it is valid).
func (s *stupidRequest) validKey(keyTable db.DBTable) bool {
	var key string
	if s.path[0] == "key" {
		if len(s.path) == 1 {
			return false
		}
		key = s.path[1]
	} else {
		k, ok := s.query["key"]
		if !ok || len(k) != 1 {
			return false
		}
		key = k[0]
	}
	err := keyTable.Get(key, &s.apiKey)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if s.apiKey.Death != -1 {
		deth := time.Unix(s.apiKey.Death, 0)
		if time.Now().After(deth) {
			return false
		}
	}
	return true
}

func (s *stupidRequest) handleKeyReq() {
	out, err := json.MarshalIndent(s.apiKey, "", "\t")
	if err != nil {
		s.w.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.w.Header().Add("content-type", "application/json")
	_, err = s.w.Write(out)
	if err != nil {
		s.w.WriteHeader(http.StatusInternalServerError)
	}
}

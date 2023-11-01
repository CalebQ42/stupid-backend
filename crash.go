package stupid

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CalebQ42/stupid-backend/crash"
	"github.com/CalebQ42/stupid-backend/db"
)

func (s *Stupid) crashReport(req *Request, table db.CrashTable) {
	if req.Method != http.MethodPost {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return
	}
	dec := json.NewDecoder(req.Body)
	var c crash.Individual
	err := dec.Decode(&c)
	req.Body.Close()
	if err != nil {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return
	}
	if f, ok := s.Apps[req.ApiKey.AppID].(CrashFilteredApp); ok && !f.AcceptCrash(c) {
		return
	}
	err = table.AddCrash(c)
	if err != nil {
		log.Printf("error while adding crash: %s", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.Resp.WriteHeader(http.StatusCreated)
}

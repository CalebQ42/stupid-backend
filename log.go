package stupid

import (
	"log"
	"net/http"
	"time"

	"github.com/CalebQ42/stupid-backend/db"
)

func (s *Stupid) logReq(req *Request, logs db.Table) {
	if req.Method != http.MethodPost {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return
	}
	id, ok := req.Query["id"]
	if !ok || len(id) != 1 {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return
	}
	plat, ok := req.Query["platform"]
	if !ok || len(plat) != 1 {
		req.Resp.WriteHeader(http.StatusBadRequest)
		return
	}
	ok, err := logs.Has(id[0])
	if err != nil {
		log.Printf("error while checking if log id is already present: %s", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	now := time.Now()
	usr := logUser{
		ID:       id[0],
		Platform: plat[0],
		LastCon:  (now.Year() * 10000) + (int(now.Month()) * 100) + (now.Day()),
	}
	if ok {
		err = logs.Update(id[0], usr)
		if err != nil {
			log.Printf("error while updating log: %s", err)
			req.Resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		_, err = logs.Add(usr)
		if err != nil {
			log.Printf("error while adding log: %s", err)
			req.Resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	req.Resp.WriteHeader(http.StatusCreated)
}

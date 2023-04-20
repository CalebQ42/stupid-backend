package stupid

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CalebQ42/stupid-backend/pkg/db"
)

func (s *Stupid) count(req *Request, tab db.Table) {
	filter := map[string]any{}
	platform := "all"
	if plat, ok := req.Query["platform"]; ok {
		if len(plat) != 1 {
			req.Resp.WriteHeader(http.StatusBadRequest)
		}
		filter = map[string]any{"platform": plat[0]}
		platform = plat[0]
	}
	i, err := tab.Count(filter)
	if err != nil {
		log.Println("Error while getting count:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	dat, err := json.Marshal(map[string]any{"platform": platform, "count": i})
	if err != nil {
		log.Println("Error while marshaling count:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = req.Resp.Write(dat)
	if err != nil {
		log.Println("Error while writing count:", err)
		req.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}
}

package stupid

import "net/http"

type Handler struct {
	userDB           string
	userCollection   string
	uploadDB         string
	uploadCollection string
	dataDB           string
	dataCollection   string
	userCount        bool
}

func (h Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {}

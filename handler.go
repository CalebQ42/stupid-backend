package stupid

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	mainDB   *mongo.Database
	usersCol *mongo.Collection
}

func (h Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "":
		fallthrough
	case "GET":
	}
}

func (h Handler) GetCapabilities(resp http.ResponseWriter) {}

func (h Handler) SendUserCount(resp http.ResponseWriter) {
	i, err := h.usersCol.EstimatedDocumentCount(context.TODO())
	if err != nil {
		fmt.Println(err)
		resp.Write([]byte("FAILURE: Can't get user count"))
		return
	}
	resp.Write([]byte(strconv.Itoa(int(i))))
}

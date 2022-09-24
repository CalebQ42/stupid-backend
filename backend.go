package stupid

import (
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

type Backend struct {
	regUsers *mongo.Collection
	client   *mongo.Client
	appIDs   []string
}

func (b Backend) init() error {
	return nil
}

func (b Backend) HandleHTTP(writer http.ResponseWriter, req *http.Request) {
	req.URL.Query()
}

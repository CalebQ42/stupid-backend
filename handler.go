package stupid

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Backend struct {
	ApiKeys     *mongo.Collection
	GlobalUsers *mongo.Collection
	AppUsers    *mongo.Collection
	AppData     *mongo.Collection
	UserData    *mongo.Collection
	Crashes     *mongo.Collection
	CountUsers  bool
}

// Implementation of http.Handler
func (b Backend) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	var features string
	if !q.Has("key") {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	} else {
		//TODO: Cache API Keys to make this a bit quicker.
		apiRes := b.ApiKeys.FindOne(context.TODO(), bson.M{"id": q.Get("key")})
		if apiRes.Err() != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			if apiRes.Err() != mongo.ErrNoDocuments {
				fmt.Println(apiRes.Err())
			}
			return
		}
		var api ApiKey
		err := apiRes.Decode(&api)
		if err != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			fmt.Println(err)
			return
		}
		if time.Unix(api.Death, 0).Before(time.Now()) {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		features = api.Features
	}
}

// Amount of users for your application within the last month.
func (b Backend) AppUserCount() (int64, error) {
	return b.AppUsers.EstimatedDocumentCount(context.TODO())
}

// Logs a connection for the given uuid.
func (b Backend) LogCon(uuid string) error {
	n := time.Now()
	time := n.Year()*10000 + int(n.Month())*100 + n.Day()
	res := b.AppUsers.FindOneAndUpdate(context.TODO(), bson.M{"_id": uuid}, bson.M{"lastConnected": time})
	if res.Err() == mongo.ErrNoDocuments {
		_, err := b.AppUsers.InsertOne(context.TODO(), bson.M{"_id": uuid, "hasGlobal": false, "lastConnected": time})
		if err != nil {
			return err
		}
	}
	return nil
}

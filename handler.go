package stupid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	var err error
	q := req.URL.Query()
	var api ApiKey
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
		err = apiRes.Decode(api)
		if err != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			fmt.Println(err)
			return
		}
		if api.Death != -1 && time.Unix(api.Death, 0).Before(time.Now()) {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	if q.Has("features") {
		err = json.NewEncoder(writer).Encode(api)
		if err != nil {
			fmt.Println(err)
			writer.WriteHeader(http.StatusFailedDependency)
		}
		return
	}
	if q.Has("logCon") {
		if !strings.Contains(api.Features, "l") {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		id := q.Get("id")
		if id != "" {
			err = b.LogCon(id)
			if err != nil {
				fmt.Println(err)
				writer.WriteHeader(http.StatusFailedDependency)
			}
		} else {
			writer.WriteHeader(http.StatusBadRequest)
		}
		return
	}
	if q.Has("userCount") {
		if !strings.Contains(api.Features, "c") {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		var count int64
		count, err = b.AppUserCount()
		if err != nil {
			fmt.Println(err)
			writer.WriteHeader(http.StatusFailedDependency)
			return
		}
		err = json.NewEncoder(writer).Encode(map[string]int64{"count": count})
		if err != nil {
			fmt.Println(err)
			writer.WriteHeader(http.StatusFailedDependency)
			return
		}
		return
	}
	writer.WriteHeader(http.StatusBadRequest)
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

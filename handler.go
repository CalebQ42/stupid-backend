package stupid

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Backend struct {
	GlobalUsers *mongo.Collection
	AppUsers    *mongo.Collection
	AppData     *mongo.Collection
	UserData    *mongo.Collection
	Crashes     *mongo.Collection
}

// Implementation of http.Handler
func (b Backend) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
	if req.Method == "GET" && req.URL.EscapedPath() == "users" {
		count, err := b.AppUserCount()
		if err != nil {
			fmt.Println(err)
		} else {
			writer.Write([]byte(strconv.Itoa(int(count))))
		}
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

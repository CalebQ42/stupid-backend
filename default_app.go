package stupid

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// A default implementation of App.
type DefaultApp struct {
	logColl   *mongo.Collection
	crashColl *mongo.Collection
	appID     string
}

// Creates a new DefaultApp and sets the log collection to appID/log and crash collection to appID/crashes.
func NewDefaultApp(appID string, client *mongo.Client) *DefaultApp {
	return &DefaultApp{
		appID:     appID,
		logColl:   client.Database(appID).Collection("log"),
		crashColl: client.Database(appID).Collection("crashes"),
	}
}

func (d DefaultApp) ID() string {
	return d.appID
}

func (d DefaultApp) Logs() *mongo.Collection {
	return d.logColl
}

func (d DefaultApp) Crashes() *mongo.Collection {
	return d.crashColl
}

func (d *DefaultApp) Initialize() error {
	return nil
}

// A DefaultApp that also can handle data requests.
type DefaultDataApp struct {
	userData *mongo.Collection
	data     *mongo.Collection
	DefaultApp
}

// Similiar to NewDefaultApp, but also adds data collections at appID/userData and appID/data.
// See /data on api.yml for how to request data.
func NewDefaultDataApp(appID string, client *mongo.Client) *DefaultDataApp {
	return &DefaultDataApp{
		DefaultApp: *NewDefaultApp(appID, client),
		userData:   client.Database(appID).Collection("userData"),
		data:       client.Database(appID).Collection("data"),
	}
}

func (d DefaultDataApp) DataRequest(r *Request) (body []byte, err error) {
	if r.Path == "/data" {
		if r.UserID == "" {
			return nil, NewStupidError(http.StatusUnauthorized)
		}
		var jsonMap map[string]any
		var dat []byte
		dat, err = io.ReadAll(r.ReqBody)
		if err != nil{
			return
		}
		err = json.Unmarshal(dat, &jsonMap)
		if err != nil {
			return
		}
		if _, ok := jsonMap["_id"]; !ok {
			return nil, NewStupidError(http.StatusBadRequest)
		}
		if _, ok := jsonMap["hint"]; !ok {
			return nil, NewStupidError(http.StatusBadRequest)
		}
		jsonMap["owner"] = r.UserID
		res := d.data.FindOneAndReplace(context.TODO(), bson.D{{Key: "_id", Value: jsonMap["_id"]}}, jsonMap)
		if res.Err() == mongo.ErrNoDocuments {
			_, err = d.data.InsertOne(context.TODO(), jsonMap)
			if err != nil {
				return
			}
		}
		return nil, NewStupidError(http.StatusCreated)
	}
	req := strings.TrimPrefix(r.Path, "/data/")
	if req == "list" {
		if r.UserID == "" {
			if !r.KeyFeatures["appData"] {
				return nil, NewStupidError(http.StatusUnauthorized)
			}
			var cur *mongo.Cursor
			if r.Query.Has("hint") {
				cur, err = d.data.Find(context.TODO(), bson.D{{Key: "hint", Value: r.Query.Get("hint")}}, options.Find().SetProjection(bson.D{{Key: "hint", Value: "1"}}))
			} else {
				cur, err = d.data.Find(context.TODO(), bson.D{{}}, options.Find().SetProjection(bson.D{{Key: "hint", Value: "1"}}))
			}
			if err != nil {
				log.Println("Err while listing data")
				return
			}
			if cur.Err() == mongo.ErrNoDocuments {
				return nil, NewStupidError(http.StatusNoContent)
			}
			var out []DataID
			err = cur.All(context.TODO(), &out)
			if err != nil {
				return
			}
			body, err = json.Marshal(out)
			return
		} else {
			if !r.KeyFeatures["userData"] {
				return nil, NewStupidError(http.StatusUnauthorized)
			}
		}
	} else if len(req) > 0 {
		if r.UserID == "" {
			if !r.KeyFeatures["appData"] {
				return nil, NewStupidError(http.StatusUnauthorized)
			}
			res := d.data.FindOne(context.TODO(), bson.D{{Key: "_id", Value: req}})
			if res.Err() == mongo.ErrNoDocuments {
				return nil, NewStupidError(http.StatusNoContent)
			} else if res.Err() != nil {
				return nil, res.Err()
			}
			var dat map[string]any
			err = res.Decode(&dat)
			if err != nil {
				return
			}
			body, err = json.Marshal(dat)
			return
		} else {
			if !r.KeyFeatures["userData"] {
				return nil, NewStupidError(http.StatusUnauthorized)
			}
			res := d.data.FindOne(context.TODO(), bson.D{{Key: "_id", Value: req}, {Key: "owner", Value: r.UserID}})
			if res.Err() == mongo.ErrNoDocuments {
				return nil, NewStupidError(http.StatusNoContent)
			} else if res.Err() != nil {
				return nil, res.Err()
			}
			var dat map[string]any
			err = res.Decode(&dat)
			if err != nil {
				return
			}
			body, err = json.Marshal(dat)
			return
		}
	}
	return nil, NewStupidError(http.StatusBadRequest)
}

// Used when asking for DefaultDataApp list.
type DataID struct {
	ID   string `bson:"_id" json:"_id"`
	Hint string `bson:"hint" json:"hint"`
}

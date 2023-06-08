package defaultapp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/CalebQ42/stupid-backend"
	"github.com/CalebQ42/stupid-backend/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// A simple implementation of stupid.App using mongodb.
type App struct {
	DB *mongo.Database
}

func NewDefaultApp(db *mongo.Database) *App {
	return &App{
		DB: db,
	}
}

func (a *App) Logs() db.Table {
	return db.NewMongoTable(a.DB.Collection("logs"))
}

func (a *App) Crashes() db.CrashTable {
	return db.NewMongoTable(a.DB.Collection("crashes"))
}

func (a *App) Extension(*stupid.Request) bool {
	return false
}

func (a *App) IgnoreOldVersionCrashes() bool {
	return false
}

func (a *App) LatestVersion() string {
	return ""
}

// A simple implementation of stupid.App using mongodb.
// Can make requests to GET: /data/{data id}?key={API key} without any
// authorization checks other then the API Key.
// Pulls from the collection "data" and data must be in the form:
/* {
	"_id": "data id",
	"data": "data",
}
*/
// The data is sent raw without any processing.
type UnauthorizedDataApp struct {
	*App
}

func NewUnauthorizedDataApp(db *mongo.Database) *UnauthorizedDataApp {
	return &UnauthorizedDataApp{
		App: NewDefaultApp(db),
	}
}

func (u *UnauthorizedDataApp) Extension(req *stupid.Request) bool {
	if len(req.Path) > 1 && req.Path[0] == "data" {
		if req.Method != http.MethodGet {
			req.Resp.WriteHeader(http.StatusBadRequest)
			return true
		}
		res := u.DB.Collection("data").FindOne(context.TODO(), bson.M{"_id": strings.Join(req.Path[1:], "/")}, options.FindOne().SetProjection(bson.M{"data": 1, "_id": 0}))
		if res.Err() == mongo.ErrNoDocuments {
			req.Resp.WriteHeader(http.StatusNotFound)
			return true
		} else if res.Err() != nil {
			log.Println("Error while finding data:", res.Err())
			req.Resp.WriteHeader(http.StatusInternalServerError)
			return true
		}
		dat := struct {
			Data string
		}{}
		err := res.Decode(&dat)
		if err != nil {
			log.Println("Error decoding data:", err)
			req.Resp.WriteHeader(http.StatusInternalServerError)
			return true
		}
		_, err = req.Resp.Write([]byte(dat.Data))
		if err != nil {
			log.Println("Error sending data:", err)
			req.Resp.WriteHeader(http.StatusInternalServerError)
		}
		return true
	}
	return false
}

// A simple implementation of stupid.App using mongodb.
// Pulls from the collection "data" and data must be in the form:
/* {
	"_id": "data id",
	"owner": "user uuid",
	"data": "data",
}
*/
// The data is sent raw without any processing.
// Requests:
/*
GET: /data/{data id}?key={API key}&token={JWT Token}
DELETE: /data/{data id}?key={API key}&token={JWT Token}
POST: /data/{data id}?key={API key}&token={JWT Token}
	Request body will be the data
GET: /data/list?key={API key}?token={JWT Token}
	Body will be a list of id strings in JSON format
*/
type AuthorizedDataApp struct {
	*App
}

func NewAuthorizedDataApp(db *mongo.Database) *AuthorizedDataApp {
	return &AuthorizedDataApp{
		App: NewDefaultApp(db),
	}
}

func (a *AuthorizedDataApp) Extension(req *stupid.Request) bool {
	if len(req.Path) > 1 && req.Path[0] == "data" {
		if req.User == nil {
			req.Resp.WriteHeader(http.StatusBadRequest)
			return true
		}
		if req.Path[1] == "list" {
			if req.Method != http.MethodGet {
				req.Resp.WriteHeader(http.StatusBadRequest)
				return true
			}
			res, err := a.DB.Collection("data").Find(context.TODO(), bson.M{"owner": req.User.ID}, options.Find().SetProjection(bson.M{"_id": 1}))
			if err == mongo.ErrNoDocuments {
				req.Resp.WriteHeader(http.StatusNotFound)
				return true
			} else if err != nil {
				log.Println("Error while getting list of data:", err)
				req.Resp.WriteHeader(http.StatusInternalServerError)
				return true
			}
			ids := make([]struct {
				ID string `bson:"_id"`
			}, 0)
			err = res.All(context.TODO(), &ids)
			if err != nil {
				log.Println("Error while decoding list:", err)
				req.Resp.WriteHeader(http.StatusInternalServerError)
				return true
			}
			out := make([]string, len(ids))
			for i := range ids {
				out[i] = ids[i].ID
			}
			outDat, err := json.Marshal(out)
			if err != nil {
				log.Println("Error while marshaling data:", err)
				req.Resp.WriteHeader(http.StatusInternalServerError)
				return true
			}
			_, err = req.Resp.Write(outDat)
			if err != nil {
				log.Println("Error while writing data:", err)
				req.Resp.WriteHeader(http.StatusInternalServerError)
			}
			return true
		}
		id := strings.Join(req.Path[1:], "/")
		switch req.Method {
		case http.MethodGet:
			res := a.DB.Collection("data").FindOne(context.TODO(), bson.M{"_id": id, "owner": req.User.ID}, options.FindOne().SetProjection(bson.M{"data": 1, "_id": 0}))
			if res.Err() == mongo.ErrNoDocuments {
				req.Resp.WriteHeader(http.StatusNotFound)
				return true
			} else if res.Err() != nil {
				log.Println("Error while getting data:", res.Err())
				req.Resp.WriteHeader(http.StatusInternalServerError)
				return true
			}
			dat := struct {
				Data string
			}{}
			err := res.Decode(&dat)
			if err != nil {
				log.Println("Error while decoding data:", err)
				req.Resp.WriteHeader(http.StatusInternalServerError)
				return true
			}
			_, err = req.Resp.Write([]byte(dat.Data))
			if err != nil {
				log.Println("Error while writing data:", err)
				req.Resp.WriteHeader(http.StatusInternalServerError)
			}
		case http.MethodDelete:
			res := a.DB.Collection("data").FindOneAndDelete(context.TODO(), bson.M{"_id": id, "owner": req.User.ID})
			if res.Err() == mongo.ErrNoDocuments {
				req.Resp.WriteHeader(http.StatusNotFound)
				return true
			} else if res.Err() != nil {
				log.Println("Error while getting data:", res.Err())
				req.Resp.WriteHeader(http.StatusInternalServerError)
			}
		case http.MethodPost:
			dat, err := io.ReadAll(req.Body)
			if err != nil {
				log.Println("Error while reading body:", err)
				req.Resp.WriteHeader(http.StatusInternalServerError)
				return true
			}
			res := a.DB.Collection("data").FindOneAndUpdate(context.TODO(), bson.M{"_id": id, "owner": req.User.ID}, bson.M{"data": string(dat)})
			if res.Err() == mongo.ErrNoDocuments {
				_, err = a.DB.Collection("data").InsertOne(context.TODO(), bson.M{"_id": id, "owner": req.User.ID, "data": string(dat)})
				if err != nil {
					log.Println("Error while creating data:", err)
					req.Resp.WriteHeader(http.StatusInternalServerError)
					return true
				}
			} else if res.Err() != nil {
				log.Println("Error while updating data:", res.Err())
				req.Resp.WriteHeader(http.StatusInternalServerError)
				return true
			}
			req.Resp.WriteHeader(http.StatusCreated)
		default:
			req.Resp.WriteHeader(http.StatusBadRequest)
		}
		return true
	}
	return false
}

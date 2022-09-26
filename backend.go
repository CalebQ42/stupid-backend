package stupid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Backend struct {
	client    *mongo.Client
	ApiKeys   *mongo.Collection
	Users     *mongo.Collection
	extension BackendExtension
	apps      map[string]App
	running   bool
}

func NewBackend(client *mongo.Client) *Backend {
	return &Backend{
		client: client,
		apps:   make(map[string]App),
	}
}

func (b *Backend) SetExtention(e BackendExtension) {
	b.extension = e
}

func (b *Backend) AddApps(app ...App) error {
	for i := range app {
		_, exist := b.apps[app[i].ID()]
		if exist {
			return errors.New("cannot add an app that already exists")
		}
		b.apps[app[i].ID()] = app[i]
	}
	if b.running {
		for i := range app {
			b.clean(app[i].ID())
		}
	}
	return nil
}

func (b *Backend) Init() {
	if b.ApiKeys == nil {
		b.ApiKeys = b.client.Database("stupid-backend").Collection("keys")
	}
	if b.Users == nil {
		b.Users = b.client.Database("stupid-backend").Collection("regUsers")
	}
	cleanTicker := time.NewTicker(time.Hour * 24)
	go func() {
		for {
			b.clean("")
			<-cleanTicker.C
		}
	}()
}

func (b Backend) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	if !query.Has("key") {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	key, err := b.GetAPIKey(query.Get("key"))
	if err == mongo.ErrNoDocuments {
		writer.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		log.Println("Err while validating api key:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if query.Has("features") {
		var out []byte
		out, err = json.Marshal(key)
		if err != nil {
			log.Println("Err while marshalling API key:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = writer.Write(out)
		if err != nil {
			log.Println("Err while sending API key:", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	app, ok := b.apps[key.AppID]
	if !ok {
		log.Println("API Key used for an app that's not set up, but API Key exists:", key.AppID)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	if query.Has("log") && key.Features["log"] {
		if !query.Has("uuid") {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		//TODO
	}
	//TODO: MORE!
	//TODO: If token is present, authenticate
	_ = app
	if b.extension != nil {
		req := Request{
			ReqBody:     req.Body,
			Query:       query,
			KeyFeatures: key.Features,
			//TODO: add UserID if authenticated.
		}
		var bod []byte
		bod, err = b.extension.HandleRequest(req)
		if err == ErrBadRequest {
			writer.WriteHeader(http.StatusBadRequest)
		} else if err == ErrUnauthorized {
			writer.WriteHeader(http.StatusUnauthorized)
		} else if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		_, err = writer.Write(bod)
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	writer.WriteHeader(http.StatusBadRequest)
}

// Creates an API Key for the given app with the given features. If features is nil, creates a key with log, registeredUsers, and sendCrash as true.
// If features is missing one of the core features (see DB.md), that feature is set to false.
func (b Backend) GenerateAPIKey(appID string, features map[string]bool, alias string) error {
	if features == nil {
		features = map[string]bool{
			"log":             true,
			"registeredUsers": true,
			"sendCrash":       true,
			"getCount":        false,
			"backend":         false,
		}
	}
	for _, feat := range []string{"log", "registeredUsers", "sendCrash", "getCount", "backend"} {
		if _, present := features[feat]; !present {
			features[feat] = false
		}
	}
	id := uuid.NewString()
	api := ApiKey{
		Key:      id,
		AppID:    appID,
		Alias:    alias,
		Death:    -1,
		Features: features,
	}
	_, err := b.ApiKeys.InsertOne(context.TODO(), api)
	return err
}

func (b Backend) GetAPIKey(key string) (out *ApiKey, err error) {
	res := b.ApiKeys.FindOne(context.TODO(), bson.D{{Key: "_id", Value: key}})
	if res.Err() != nil {
		return nil, res.Err()
	}
	out = new(ApiKey)
	err = res.Decode(out)
	return
}

func (b Backend) clean(id string) (err error) {
	if id == "" {
		for i := range b.apps {
			err = b.clean(i)
			if err != nil {
				return
			}
		}
		return
	}
	now := time.Now()
	dayInt := int(now.Year())*10000 + int(now.Month())*100 + int(now.Day())
	filter := options.Find().SetMin(bson.D{{Key: "lastConnected", Value: dayInt}})
	res, err := b.apps[id].Logs().Find(context.TODO(), bson.D{}, filter) //TODO: stuff
	if err != nil && err != mongo.ErrNoDocuments {
		return
	}
	fmt.Println(res.Current)
	return
}

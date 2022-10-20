package stupid

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
			err := b.clean("")
			if err != nil {
				log.Println("Error while cleaning:", err)
			}
			<-cleanTicker.C
		}
	}()
}

// Creates an API Key for the given app with the given features. If features is nil, creates a key with log, registeredUsers, and sendCrash as true.
// If features is missing one of the core features (see DB.md), that feature is set to false.
func (b Backend) GenerateAPIKey(appID string, features map[string]bool, alias string) error {
	if features == nil {
		features = map[string]bool{
			"log":             true,
			"registeredUsers": true,
			"sendCrash":       true,
			"backend":         false,
		}
	}
	for _, feat := range []string{"log", "registeredUsers", "sendCrash", "backend"} {
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

func (b Backend) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if req.Body != nil {
		defer req.Body.Close()
	}
	query := req.URL.Query()
	if !query.Has("key") {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	path := req.URL.Path
	key, err := b.GetAPIKey(query.Get("key"))
	if err == mongo.ErrNoDocuments {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Println("Err while validating api key:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if path == "/features" {
		err = retMarshallable(key, writer)
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
	r := &Request{
		ReqBody:     req.Body,
		Method:      req.Method,
		Query:       query,
		KeyFeatures: key.Features,
		Path:        path,
	}
	if authApp, ok := app.(AuthenticatedDataApp); ok && query.Has("token") {
		r.UserID = verifyToken(authApp, query.Get("token"))
	}
	if dataApp, ok := app.(DataApp); ok && strings.HasPrefix(path, "/data") {
		var body []byte
		body, err = dataApp.DataRequest(r)
		if stupidErr, ok := err.(StupidError); ok {
			writer.WriteHeader(stupidErr.code)
			return
		} else if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = writer.Write(body)
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	switch path {
	case "/log":
		handleLog(writer, app, r)
	case "/crash":
		handleCrash(writer, app, r)
	case "/createUser":
		b.createUser(writer, app, r)
	case "/auth":
		b.authUser(writer, app, r)
	default:
		if b.extension != nil {
			var bod []byte
			bod, err = b.extension.HandleRequest(r)
			if stupidErr, ok := err.(StupidError); ok {
				writer.WriteHeader(stupidErr.code)
				return
			} else if err != nil {
				log.Println(err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = writer.Write(bod)
			if err != nil {
				log.Println(err)
				writer.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			writer.WriteHeader(http.StatusBadRequest)
		}
	}
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
	toDel := time.Now().Add(time.Duration(-30*24) * time.Hour)
	dayInt := int(toDel.Year())*10000 + int(toDel.Month())*100 + int(toDel.Day())
	_, err = b.apps[id].Logs().DeleteMany(context.TODO(), bson.D{{Key: "lastConn", Value: bson.D{{Key: "$lt", Value: dayInt}}}}) //TODO: stuff
	if err != nil {
		return
	}
	return
}

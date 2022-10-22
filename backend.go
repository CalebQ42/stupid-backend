package stupid

import (
	"context"
	"crypto/ed25519"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Actual Stupid-Backend. Implements http.Handler.
type Backend struct {
	extension BackendExtension
	apiKeys   *mongo.Collection
	users     *mongo.Collection
	apps      map[string]App
	pubKey    ed25519.PublicKey
	privKey   ed25519.PrivateKey
	running   bool
}

// Creates a new Backend and sets they ApiKey collection to stupid-backend/keys.
func NewBackendFromClient(client *mongo.Client) *Backend {
	return &Backend{
		apiKeys: client.Database("stupid-backend").Collection("keys"),
		apps:    make(map[string]App),
	}
}

// Creates a new Backend using the given collections for ApiKeys.
func NewBackend(apiKeys *mongo.Collection) *Backend {
	return &Backend{
		apiKeys: apiKeys,
		apps:    make(map[string]App),
	}
}

// Add the ability to authenticate users. Suggested collection to use: stupid-backend/users.
func (b *Backend) AddUsers(users *mongo.Collection, pub ed25519.PublicKey, priv ed25519.PrivateKey) {
	b.users = users
	b.pubKey = pub
	b.privKey = priv
}

// Adds a BackendExtension to further extend functionality.
func (b *Backend) SetExtention(e BackendExtension) {
	b.extension = e
}

// Adds the given stupid.App to the backend.
// If the app uses
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
			err := b.clean(app[i].ID())
			if err != nil {
				log.Println("Err while cleaning", i, ":", err)
			}
		}
	}
	return nil
}

// Initialize individual apps and starts cleanup of the logs.
// Logs are cleaned every 24 hours.
func (b *Backend) Init() error {
	for _, app := range b.apps {
		err := app.Initialize()
		if err != nil {
			return err
		}
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
	return nil
}

// Creates an API Key for the given app with the given features. If features is nil, creates a key with log, sendCrash, appData, and userData as true, with registeredUsers set to true if user collection has been populated.
// If features is missing one of the core features (see DB.md not including data features), that feature is set to false.
func (b Backend) GenerateAPIKey(appID string, features map[string]bool, alias string) error {
	if features == nil {
		features = map[string]bool{
			"log":             true,
			"registeredUsers": b.users != nil,
			"sendCrash":       true,
			"appData":         true,
			"userData":        true,
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
	_, err := b.apiKeys.InsertOne(context.TODO(), api)
	return err
}

func (b Backend) GetAPIKey(key string) (out *ApiKey, err error) {
	res := b.apiKeys.FindOne(context.TODO(), bson.D{{Key: "_id", Value: key}})
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
	if b.users != nil && query.Has("token") {
		r.UserID = b.verifyToken(query.Get("token"))
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

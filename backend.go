package stupid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Backend struct {
	client  *mongo.Client
	ApiKeys *mongo.Collection
	Users   *mongo.Collection
	apps    map[string]App
	running bool
}

func NewBackend(client *mongo.Client) *Backend {
	return &Backend{
		client: client,
	}
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
	fmt.Println("yo")
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
	res, err := b.apps[id].Logs().Find(context.TODO(), bson.D{}, filter)
	if err != nil && err != mongo.ErrNoDocuments {
		return
	}
	fmt.Println(res.Current)
	return
}

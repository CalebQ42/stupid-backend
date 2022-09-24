package stupid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Backend struct {
	client  *mongo.Client
	apps    map[string]App
	running bool
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

func (b Backend) clean(id string) (err error) {
	if id == "" {
		for i := range b.apps {
			err = b.clean(i)
			if err != nil {
				return
			}
		}
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

func (b Backend) Init() {
	cleanTicker := time.NewTicker(time.Hour * 24)
	go func() {
		for {
			<-cleanTicker.C
			b.clean("")
		}
	}()
}

func (b Backend) HandleHTTP(writer http.ResponseWriter, req *http.Request) {
	req.URL.Query()
}

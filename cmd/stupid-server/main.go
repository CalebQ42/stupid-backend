package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/CalebQ42/stupid-backend"
	"github.com/CalebQ42/stupid-backend/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	mongoCon := os.Getenv("MONGO")
	if mongoCon == "" {
		panic("$MONGO must be set")
	}
	cl, err := mongo.NewClient(options.Client().ApplyURI(mongoCon))
	if err != nil {
		panic(fmt.Errorf("can't connect to mongo client: %s", err))
	}
	err = cl.Connect(context.TODO())
	if err != nil {
		panic(fmt.Errorf("can't connect to mongo client: %s", err))
	}
	keyTab := db.NewMongoTable(cl.Database("stupid-backend").Collection("keys"))
	st := stupid.NewStupidBackend(keyTab)
	st.AppTables = func(id string) db.App {
		return db.App{
			Logs:    db.NewMongoTable(cl.Database(id).Collection("log")),
			Crashes: db.NewMongoTable(cl.Database(id).Collection("crashes")),
		}
	}
	http.ListenAndServe(":4223", st)
}

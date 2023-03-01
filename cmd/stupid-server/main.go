package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/CalebQ42/stupid-backend/pkg/db"
	"github.com/CalebQ42/stupid-backend/pkg/stupid"
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
	st := &stupid.Stupid{
		Keys: keyTab,
	}
	http.ListenAndServe(":4223", st)
}

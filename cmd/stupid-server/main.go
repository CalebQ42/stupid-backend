package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/CalebQ42/stupid-backend"
	"github.com/CalebQ42/stupid-backend/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func usage() {
	fmt.Println("USAGE: ./stupid-server -key [key file] -cert [certificate file] [MongoDB connection string] [listen address]")
	fmt.Println("")
	fmt.Println("MongoDB connection string MUST be provided.")
	fmt.Println("If key or cert files are not provided, the server will be started with http.")
	fmt.Println("If listen address is not provided, \":4223\" is used (localhost:4223).")
}

func main() {
	flag.Usage = usage
	keyStr := flag.String("key", "", "Key file")
	certStr := flag.String("cert", "", "cert File")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Must provide MongoDB connection URL.")
	}
	cl, err := mongo.NewClient(options.Client().ApplyURI(args[0]))
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
	listen := ":4223"
	if len(args) > 1 {
		listen = args[1]
	}
	if *keyStr == "" || *certStr == "" {
		err = http.ListenAndServe(listen, st)
	} else {
		err = http.ListenAndServeTLS(listen, *certStr, *keyStr, st)
	}
	if err != nil {
		fmt.Println(err)
	}
}

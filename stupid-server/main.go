package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/CalebQ42/stupid-backend"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func help() {
	fmt.Println("Usage: stupid -tlsdir <tls key location> <mongodb connection string>")
	fmt.Println()
	flag.PrintDefaults()
}

func main() {
	appList := flag.String("apps", "testing", "Comma deliniated list of apps to use. If API Keys are not already created, new keys are created.")
	addr := flag.String("addr", ":4223", "Address to open the server on.")
	keysDir := flag.String("tlsdir", "", "Directory with key.pem and cert.pem. Defaults to $HOME. Required.")
	flag.Usage = help
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Please provied MongoDB connection string")
		os.Exit(1)
	}
	if !strings.HasPrefix(args[0], "mongodb://") {
		fmt.Println("Provided MongoDB connection string is not a Mongo address")
		os.Exit(1)
	}
	var err error
	if *keysDir == "" {
		*keysDir, err = os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(args[0]))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	appIDs := strings.Split(*appList, ",")
	for i := range appIDs {
		appIDs[i] = strings.TrimSpace(appIDs[i])
	}
	backend := stupid.NewBackend(client)
	for i := range appIDs {
		err = backend.AddApps(stupid.NewDefaultApp(appIDs[i], client))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	backend.Init()

	for i := range appIDs {
		var res *mongo.Cursor
		res, err = backend.ApiKeys.Find(context.TODO(), bson.D{{Key: "appID", Value: appIDs[i]}})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var keys []stupid.ApiKey
		err = res.All(context.TODO(), &keys)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if len(keys) > 0 {
			continue
		}
		err = backend.GenerateAPIKey(appIDs[i], nil, "default key for "+appIDs[i])
		fmt.Println("Generating key for " + appIDs[i])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	err = http.ListenAndServeTLS(*addr, path.Join(*keysDir, "cert.pem"), path.Join(*keysDir, "key.pem"), backend)
	fmt.Println(err)
}

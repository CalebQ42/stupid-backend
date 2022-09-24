package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/CalebQ42/stupid-backend"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func help() {
	fmt.Println("Usage: stupid -tlsdir <tls key location> <mongodb connection string>")
	flag.PrintDefaults()
}

func main() {
	appList := flag.String("apps", "testing", "Comma deliniated list of apps to use. If none given, uses \"testing\". If API Keys are not already created, new keys are created.")
	port := flag.Int("port", 4223, "Port to open requests on. Defaults to 4223.")
	keysDir := flag.String("tlsdir", "", "Directory with key.pem and cert.pem. Defaults to $HOME. Required.")
	flag.Usage = help
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		panic("Please provied MongoDB connection string")
	}
	if !strings.HasPrefix(args[0], "mongo://") {
		panic("Provided MongoDB connection string is not a Mongo address")
	}
	var err error
	if *keysDir == "" {
		*keysDir, err = os.UserHomeDir()
		if err != nil {
			panic(err)
		}
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(args[0]))
	if err != nil {
		panic(err)
	}
	appIDs := strings.Split(*appList, ",")
	for i := range appIDs {
		appIDs[i] = strings.TrimSpace(appIDs[i])
	}
	apps := make([]*stupid.DefaultApp, len(appIDs))
	for i := range apps {
		apps[i] = stupid.NewDefaultApp(appIDs[i], client)
	}
	backend := stupid.NewBackend(client)
	backend.Init()
	//TODO: check for keys and create new ones.
	err = http.ListenAndServeTLS(":"+strconv.Itoa(*port), path.Join(*keysDir, "cert.pem"), path.Join(*keysDir, "key.pem"), backend)
	fmt.Println(err)
}

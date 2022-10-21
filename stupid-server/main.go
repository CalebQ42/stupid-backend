package main

import (
	"context"
	"crypto/ed25519"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/CalebQ42/stupid-backend"
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
	keysDir := flag.String("tlsdir", "", "Directory with key.pem and cert.pem. Defaults to $PWD. Required.")
	tokenDir := flag.String("keydir", "", "Directory with stupid-pub.key and stupid-priv.key for registered users (will be generated if not present). EdDSA keys. If present, allows for user authentication.")
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
		*keysDir, err = os.Getwd()
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
	backend := stupid.NewBackendFromClient(client)
	if *tokenDir != "" {
		var genKeys bool
		pubLoc, privLoc := path.Join(*tokenDir, "stupid-pub.key"), path.Join(*tokenDir, "stupid-priv.key")
		var pub, priv *os.File
		pub, err = os.Open(pubLoc)
		if os.IsNotExist(err) {
			genKeys = true
		} else if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		priv, err = os.Open(privLoc)
		if os.IsNotExist(err) {
			genKeys = true
		} else if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var pu ed25519.PublicKey
		var pr ed25519.PrivateKey
		if genKeys {
			os.Rename(pubLoc, pubLoc+".bak")
			os.Rename(privLoc, privLoc+".bak")
			pub, err = os.Create(pubLoc)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			priv, err = os.Create(privLoc)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			pu, pr, err = ed25519.GenerateKey(nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			_, err = pub.Write(pu)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			_, err = priv.Write(pr)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			backend.AddUsers(client.Database("stupid-backend").Collection("users"), pu, pr)
		} else {
			var pubKey, privKey []byte
			pubKey, err = io.ReadAll(pub)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			privKey, err = io.ReadAll(priv)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			pu = pubKey
			pr = privKey
		}
		backend.AddUsers(client.Database("stupid-backend").Collection("users"), pu, pr)
	}
	for i := range appIDs {
		//TODO: Move to DefaultDataApp.
		err = backend.AddApps(stupid.NewDefaultDataApp(appIDs[i], client))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	backend.Init()
	err = http.ListenAndServeTLS(*addr, path.Join(*keysDir, "cert.pem"), path.Join(*keysDir, "key.pem"), backend)
	fmt.Println(err)
}

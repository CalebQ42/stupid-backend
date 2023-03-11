package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/CalebQ42/stupid-backend"
	"github.com/CalebQ42/stupid-backend/pkg/db"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func usage() {
	fmt.Println("USAGE: ./stupid-server -key [key file] -cert [certificate file] -userpub [user public key] -userpriv [user private key] [MongoDB connection string] [listen address]")
	fmt.Println("")
	fmt.Println("MongoDB connection string MUST be provided.")
	fmt.Println("If key or cert files are not provided, the server will be started with http and user authentication will be disabled.")
	fmt.Println("If userpub and userpriv is given, but do not exist the keys will be created.")
	fmt.Println("If listen address is not provided, \":4223\" is used (localhost:4223).")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	keyStr := flag.String("key", "", "Key file")
	certStr := flag.String("cert", "", "cert File")
	userPub := flag.String("userpub", "", "User public key")
	userPriv := flag.String("userpriv", "", "User private key")
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
	https := *keyStr != "" && *certStr != ""
	if *userPub != "" && *userPriv != "" {
		if !https {
			panic(errors.New("user keys provided, but not TLS key and certificate"))
		}
		privNotExist, pubNotExist := false, false
		var pub, priv *os.File
		pub, err = os.Open(*userPub)
		if os.IsNotExist(err) {
			pubNotExist = true
		} else if err != nil {
			panic(fmt.Errorf("can't open user public key: %w", err))
		}
		priv, err = os.Open(*userPriv)
		if os.IsNotExist(err) {
			privNotExist = true
		} else if err != nil {
			panic(fmt.Errorf("can't open user private key: %w", err))
		}
		if privNotExist != pubNotExist {
			panic(errors.New("only user private or public key exists and not both or neither"))
		}
		var pubKey ed25519.PublicKey
		var privKey ed25519.PrivateKey
		if privNotExist {
			pubKey, privKey, err = ed25519.GenerateKey(rand.Reader)
			if err != nil {
				panic(fmt.Errorf("can't create user public and private keys: %w", err))
			}
			pub, err = os.Create(*userPub)
			if err != nil {
				panic(fmt.Errorf("can't create user public key file: %w", err))
			}
			priv, err = os.Create(*userPriv)
			if err != nil {
				panic(fmt.Errorf("can't create user private key file: %w", err))
			}
			_, err = pub.Write(pubKey)
			if err != nil {
				panic(fmt.Errorf("can't write user public key: %w", err))
			}
			_, err = priv.Write(privKey)
			if err != nil {
				panic(fmt.Errorf("can't write user private key: %w", err))
			}
		} else {
			pubKey, err = io.ReadAll(pub)
			if err != nil {
				panic(fmt.Errorf("can't read public key: %w", err))
			}
			privKey, err = io.ReadAll(priv)
			if err != nil {
				panic(fmt.Errorf("can't read private key: %w", err))
			}
		}
		fmt.Println("User authentication enabled!")
		userTab := db.NewMongoTable(cl.Database("stupid-backend").Collection("users"))
		st.EnableUserAuth(userTab, pubKey, privKey)
	}
	listen := ":4223"
	if len(args) > 1 {
		listen = args[1]
	}
	if https {
		fmt.Println("Starting HTTPS server")
		err = http.ListenAndServeTLS(listen, *certStr, *keyStr, st)
	} else {
		fmt.Println("Starting HTTP server")
		err = http.ListenAndServe(listen, st)
	}
	if err != nil {
		fmt.Println(err)
	}
}

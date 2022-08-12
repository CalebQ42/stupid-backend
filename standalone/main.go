package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/CalebQ42/stupid-backend"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	flag.Usage = helpMsg
	mainDB := flag.String("db", "stupid", "The main DB to use. Unless specified, uses the DB named stupid")
	globalUs := flag.String("glob", "globalUsers", "Global users collection. Can specify a DB outside of the main by using /. Ex: stupid/globalUsers")
	appUs := flag.String("users", "users", "App users collection. Can specify a DB outside of the main by using /. Ex: stupid/users")
	appData := flag.String("appData", "appData", "App data collection. Can specify a DB outside of the main by using /. Ex: stupid/appData")
	userData := flag.String("data", "userData", "User data collection. Can specify a DB outside of the main by using /. Ex: stupid/userData")
	crash := flag.String("crash", "crash", "Crashes collection. Can specify a DB outside of the main by using /. Ex: stupid/crash")
	api := flag.String("api", "api", "API key collection. Can specify a DB outside of the main by using /. Ex: stupid/api")
	flag.Parse()
	if flag.Arg(0) == "" {
		fmt.Println("Must provide MongoDB address")
		os.Exit(1)
	}
	mong, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(flag.Arg(0)))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var globCol, appUserCol, appDatCol, userDatCol, crashCol, apiCol *mongo.Collection
	if !strings.Contains(*globalUs, "/") {
		globCol = mong.Database(*mainDB).Collection(*globalUs)
	} else {
		sep := strings.Index(*globalUs, "/")
		globCol = mong.Database((*globalUs)[:sep]).Collection((*globalUs)[sep+1:])
	}
	if !strings.Contains(*appUs, "/") {
		appUserCol = mong.Database(*mainDB).Collection(*appUs)
	} else {
		sep := strings.Index(*appUs, "/")
		appUserCol = mong.Database((*appUs)[:sep]).Collection((*appUs)[sep+1:])
	}
	if !strings.Contains(*appData, "/") {
		appDatCol = mong.Database(*mainDB).Collection(*appData)
	} else {
		sep := strings.Index(*appData, "/")
		appDatCol = mong.Database((*appData)[:sep]).Collection((*appData)[sep+1:])
	}
	if !strings.Contains(*userData, "/") {
		userDatCol = mong.Database(*mainDB).Collection(*userData)
	} else {
		sep := strings.Index(*userData, "/")
		userDatCol = mong.Database((*userData)[:sep]).Collection((*userData)[sep+1:])
	}
	if !strings.Contains(*crash, "/") {
		crashCol = mong.Database(*mainDB).Collection(*crash)
	} else {
		sep := strings.Index(*crash, "/")
		crashCol = mong.Database((*crash)[:sep]).Collection((*crash)[sep+1:])
	}
	if !strings.Contains(*api, "/") {
		apiCol = mong.Database(*mainDB).Collection(*api)
	} else {
		sep := strings.Index(*api, "/")
		apiCol = mong.Database((*api)[:sep]).Collection((*api)[sep+1:])
	}
	http.ListenAndServe(":1109", stupid.Backend{
		GlobalUsers: globCol,
		AppUsers:    appUserCol,
		AppData:     appDatCol,
		UserData:    userDatCol,
		Crashes:     crashCol,
		ApiKeys:     apiCol,
	})
}

func helpMsg() {
	fmt.Println("Usage: stupid [mongo address] -db [value] -glob [value] -users [value] -appData [value] -data [value] -crash [value] -api [value]")
	fmt.Println("")
	fmt.Println("I'm too lazy ATM (plus things are probably going to change) to make a proper help message. I'll do it later.")
}

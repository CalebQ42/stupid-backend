package stupid

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (b Backend) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if req.Body != nil {
		defer req.Body.Close()
	}
	query := req.URL.Query()
	if !query.Has("key") {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	key, err := b.GetAPIKey(query.Get("key"))
	if err == mongo.ErrNoDocuments {
		writer.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		log.Println("Err while validating api key:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if query.Has("features") {
		var out []byte
		out, err = json.Marshal(key)
		if err != nil {
			log.Println("Err while marshalling API key:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = writer.Write(out)
		if err != nil {
			log.Println("Err while sending API key:", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	app, ok := b.apps[key.AppID]
	if !ok {
		log.Println("API Key used for an app that's not set up, but API Key exists:", key.AppID)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	if query.Has("log") {
		if !key.Features["log"] {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !query.Has("uuid") || !query.Has("plat") {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		now := time.Now()
		newDate := int(now.Year())*10000 + int(now.Month())*100 + int(now.Day())
		res := app.Logs().FindOneAndUpdate(context.TODO(), bson.D{{Key: "_id", Value: query.Get("uuid")}}, bson.D{{Key: "$set", Value: bson.D{{Key: "lastConn", Value: newDate}, {Key: "plat", Value: query.Get("plat")}}}})
		if res.Err() == mongo.ErrNoDocuments {
			_, err = app.Logs().InsertOne(context.TODO(), bson.D{{Key: "_id", Value: query.Get("uuid")}, {Key: "lastConn", Value: newDate}, {Key: "plat", Value: query.Get("plat")}})
			if err != nil {
				log.Println("Err while logging:", err)
				writer.WriteHeader(http.StatusInternalServerError)
			}
		} else if res.Err() != nil {
			log.Println("Err while logging:", res.Err())
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if query.Has("crash") {
		if !key.Features["sendCrash"] {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		var crash IndCrash
		var dat []byte
		dat, err = io.ReadAll(req.Body)
		if err != nil {
			log.Println("Err while reading body:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(dat, &crash)
		if err != nil {
			log.Println("Err while marshalling crash report:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		firstLine, _, _ := strings.Cut(crash.Stack, "\n")
		res := app.Crashes().FindOneAndUpdate(context.TODO(), bson.D{{Key: "err", Value: crash.Err}, {Key: "first", Value: firstLine}}, bson.D{{Key: "$addToSet", Value: bson.D{{Key: "crashes", Value: crash}}}})
		if res.Err() == mongo.ErrNoDocuments {
			newGroup := GroupCrash{
				ID:        uuid.NewString(),
				Err:       crash.Err,
				FirstLine: firstLine,
				Crashes: []IndCrash{
					crash,
				},
			}
			_, err = app.Crashes().InsertOne(context.TODO(), newGroup)
			if err != nil {
				log.Println("Err while creating new crash group:", err)
				writer.WriteHeader(http.StatusInternalServerError)
			}
		} else if res.Err() != nil {
			log.Println("Err while reporting crash:", res.Err())
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	//TODO: Authenticate!
	//TODO: If token is present, authenticate
	if b.extension != nil {
		req := Request{
			ReqBody:     req.Body,
			Query:       query,
			KeyFeatures: key.Features,
			//TODO: add UserID if authenticated.
		}
		var bod []byte
		bod, err = b.extension.HandleRequest(req)
		if err == ErrBadRequest {
			writer.WriteHeader(http.StatusBadRequest)
		} else if err == ErrUnauthorized {
			writer.WriteHeader(http.StatusUnauthorized)
		} else if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		_, err = writer.Write(bod)
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	writer.WriteHeader(http.StatusBadRequest)
}

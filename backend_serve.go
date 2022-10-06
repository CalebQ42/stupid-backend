package stupid

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/argon2"
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
		err = retMarshallable(key, writer)
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
		res := app.Crashes().FindOne(context.TODO(), bson.D{{Key: "crashes._id", Value: crash.ID}})
		if res.Err() == nil {
			return
		} else if res.Err() != mongo.ErrNoDocuments {
			log.Println("Err while finding existing crash report:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		firstLine, _, _ := strings.Cut(crash.Stack, "\n")
		res = app.Crashes().FindOneAndUpdate(context.TODO(), bson.D{{Key: "err", Value: crash.Err}, {Key: "first", Value: firstLine}}, bson.D{{Key: "$addToSet", Value: bson.D{{Key: "crashes", Value: crash}}}})
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
	if query.Has("createUser") {
		if !key.Features["registeredUsers"] {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		if req.Method != http.MethodPost {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		var dat []byte
		dat, err = io.ReadAll(req.Body)
		if err != nil {
			log.Println("Err while reading body for create:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(dat) == 0 {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		var body CreateRequest
		err = json.Unmarshal(dat, &body)
		if err != nil || body.Password == "" || body.Username == "" || body.Email == "" {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		var ret CreateReturn
		res := b.Users.FindOne(context.TODO(), bson.D{{Key: "username", Value: body.Username}})
		if res.Err() == nil {
			ret.Problem = "username"
			err = retMarshallable(ret, writer)
			if err != nil {
				log.Println("Err while returning create request:", err)
			}
			return
		} else if res.Err() != mongo.ErrNoDocuments {
			log.Println("Err while returning create request:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(body.Password) > 32 || len(body.Password) < 5 {
			ret.Problem = "password"
			err = retMarshallable(ret, writer)
			if err != nil {
				log.Println("Err while returning create request:", err)
			}
			return
		}
		salt := make([]byte, 16)
		_, err = rand.Reader.Read(salt)
		if err != nil {
			log.Println("Err while generating salt:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		newUser := RegUser{
			ID:       uuid.NewString(),
			Username: body.Username,
			Password: base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(body.Password), salt, 1, 64*1024, 4, 32)),
			Salt:     base64.RawStdEncoding.EncodeToString(salt),
			Email:    body.Email,
		}
		_, err = b.Users.InsertOne(context.TODO(), newUser)
		if err != nil {
			log.Println("Err while inserting new user:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		ret.ID = newUser.ID
		//TODO: generate jwt token
		err = retMarshallable(ret, writer)
		if err != nil {
			log.Println("Err while inserting new user:", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if query.Has("auth") {
		if !key.Features["registeredUsers"] {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		if req.Method != http.MethodPost {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		var dat []byte
		dat, err = io.ReadAll(req.Body)
		if err != nil {
			log.Println("Err while reading body for auth:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(dat) == 0 {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		var body AuthRequest
		err = json.Unmarshal(dat, &body)
		if err != nil || body.Password == "" || body.Username == "" {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		var ret AuthReturn
		res := b.Users.FindOne(context.TODO(), bson.D{{Key: "username", Value: body.Username}})
		if res.Err() == mongo.ErrNoDocuments {
			err = retMarshallable(ret, writer)
			if err != nil {
				log.Println("Err while returning auth request:", err)
			}
			return
		} else if res.Err() != nil {
			log.Println("Err while checking for registered user:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		var user RegUser
		err = res.Decode(&user)
		if err != nil {
			log.Println("Err while decoding registered user:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		//TODO: timeout.
		body.Password = string(argon2.IDKey([]byte(body.Password), []byte(user.Salt), 1, 64*1024, 4, 32))
		if base64.RawStdEncoding.EncodeToString([]byte(body.Password)) != user.Password {
			//TODO: increment failed
			return
		}
		ret.ID = user.ID
		//TODO: Generate jwt token
	}
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

func retMarshallable(m any, w io.Writer) (err error) {
	out, err := json.Marshal(m)
	if err != nil {
		return
	}
	_, err = w.Write(out)
	return
}
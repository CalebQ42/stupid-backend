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

func handleLog(writer http.ResponseWriter, app App, r *Request) {
	if !r.KeyFeatures["log"] {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !r.Query.Has("uuid") || !r.Query.Has("plat") {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	now := time.Now()
	newDate := int(now.Year())*10000 + int(now.Month())*100 + int(now.Day())
	res := app.Logs().FindOneAndUpdate(context.TODO(), bson.D{{Key: "_id", Value: r.Query.Get("uuid")}}, bson.D{{Key: "$set", Value: bson.D{{Key: "lastConn", Value: newDate}, {Key: "plat", Value: r.Query.Get("plat")}}}})
	if res.Err() == mongo.ErrNoDocuments {
		_, err := app.Logs().InsertOne(context.TODO(), bson.D{{Key: "_id", Value: r.Query.Get("uuid")}, {Key: "lastConn", Value: newDate}, {Key: "plat", Value: r.Query.Get("plat")}})
		if err != nil {
			log.Println("Err while logging:", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
	} else if res.Err() != nil {
		log.Println("Err while logging:", res.Err())
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func handleCrash(writer http.ResponseWriter, app App, r *Request) {
	if !r.KeyFeatures["sendCrash"] {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	var crash IndCrash
	dat, err := io.ReadAll(r.ReqBody)
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
}

func (b Backend) createUser(writer http.ResponseWriter, app App, r *Request) {
	if !r.KeyFeatures["registeredUsers"] {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	if b.users == nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	dat, err := io.ReadAll(r.ReqBody)
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
	res := b.users.FindOne(context.TODO(), bson.D{{Key: "username", Value: body.Username}})
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
	newUser := User{
		ID:       uuid.NewString(),
		Username: body.Username,
		Password: base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(body.Password), salt, 1, 64*1024, 4, 32)),
		Salt:     base64.RawStdEncoding.EncodeToString(salt),
		Email:    body.Email,
	}
	_, err = b.users.InsertOne(context.TODO(), newUser)
	if err != nil {
		log.Println("Err while inserting new user:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	ret.ID = newUser.ID
	var token []byte
	token, err = b.createToken(newUser.ID)
	if err != nil {
		log.Println("Err while creating token:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	ret.Token = string(token)
	err = retMarshallable(ret, writer)
	if err != nil {
		log.Println("Err while inserting new user:", err)
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func (b Backend) authUser(writer http.ResponseWriter, app App, r *Request) {
	if !r.KeyFeatures["registeredUsers"] {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	if b.users == nil {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	dat, err := io.ReadAll(r.ReqBody)
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
	res := b.users.FindOne(context.TODO(), bson.D{{Key: "username", Value: body.Username}})
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
	var user User
	err = res.Decode(&user)
	if err != nil {
		log.Println("Err while decoding registered user:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user.Failed > 2 {
		if user.LastTimeout == 0 {
			_, err = b.users.UpdateByID(context.TODO(), bson.D{{Key: "_id", Value: user.ID}}, bson.D{{Key: "$set", Value: bson.D{{Key: "lastTimeout", Value: time.Now().Unix()}}}})
			if err != nil {
				log.Println("Err while updating lastTimout:", err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			ret.Timeout = 3 ^ ((user.Failed / 3) - 1)
			err = retMarshallable(ret, writer)
			if err != nil {
				log.Println("Err while authenticating user:", err)
				writer.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		var timeoutTime time.Time
		if user.Failed > 18 {
			timeoutTime = time.Unix(int64(user.LastTimeout), 0).Add(time.Duration(3^5) * time.Second)
		} else {
			timeoutTime = time.Unix(int64(user.LastTimeout), 0).Add(time.Duration(3^((user.Failed/3)-1)) * time.Second)
		}
		if timeoutTime.After(time.Now()) {
			ret.Timeout = int(time.Until(timeoutTime).Seconds())
			err = retMarshallable(ret, writer)
			if err != nil {
				log.Println("Err while authenticating user:", err)
				writer.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}
	body.Password = string(argon2.IDKey([]byte(body.Password), []byte(user.Salt), 1, 64*1024, 4, 32))
	if base64.RawStdEncoding.EncodeToString([]byte(body.Password)) != user.Password {
		user.Failed++
		if user.Failed%3 == 0 {
			_, err = b.users.UpdateByID(context.TODO(), bson.D{{Key: "_id", Value: user.ID}}, bson.D{{Key: "$set", Value: bson.D{{Key: "failed", Value: user.Failed}, {Key: "lastTimeout", Value: time.Now().Unix()}}}})
			if err != nil {
				log.Println("Err while updating failed:", err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			if user.Failed > 18 {
				ret.Timeout = 3 ^ 5
			} else {
				ret.Timeout = 3 ^ ((user.Failed / 3) - 1)
			}
		} else {
			_, err = b.users.UpdateByID(context.TODO(), bson.D{{Key: "_id", Value: user.ID}}, bson.D{{Key: "$inc", Value: bson.D{{Key: "failed", Value: 1}}}})
			if err != nil {
				log.Println("Err while updating failed:", err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		err = retMarshallable(ret, writer)
		if err != nil {
			log.Println("Err while authenticating user:", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	ret.ID = user.ID
	var token []byte
	token, err = b.createToken(user.ID)
	if err != nil {
		log.Println("Err while creating token:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	ret.Token = string(token)
	err = retMarshallable(ret, writer)
	if err != nil {
		log.Println("Err while authenticating user:", err)
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func retMarshallable(m any, w io.Writer) (err error) {
	out, err := json.Marshal(m)
	if err != nil {
		return
	}
	_, err = w.Write(out)
	return
}

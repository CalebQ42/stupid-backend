package stupid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/argon2"
)

type Backend struct {
	ApiKeys     *mongo.Collection
	GlobalUsers *mongo.Collection
	AppUsers    *mongo.Collection
	AppData     *mongo.Collection
	UserData    *mongo.Collection
	Crashes     *mongo.Collection
	CountUsers  bool
}

// Implementation of http.Handler
func (b Backend) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	var err error
	q := req.URL.Query()
	var api ApiKey
	if !q.Has("key") {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	} else {
		//TODO: Cache API Keys to make this a bit quicker.
		apiRes := b.ApiKeys.FindOne(context.TODO(), bson.M{"": q.Get("key")})
		if apiRes.Err() != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			if apiRes.Err() != mongo.ErrNoDocuments {
				fmt.Println(apiRes.Err())
			}
			return
		}
		err = apiRes.Decode(&api)
		if err != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			fmt.Println(err)
			return
		}
		if api.Death != -1 && time.Unix(api.Death, 0).Before(time.Now()) {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	var body map[string]interface{}
	if req.ContentLength > 0 {
		err = json.NewDecoder(req.Body).Decode(&body)
		if err != nil {
			fmt.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		body = make(map[string]interface{})
	}
	if q.Has("features") {
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "	")
		err = enc.Encode(api)
		if err != nil {
			fmt.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if q.Has("logCon") {
		if !strings.Contains(api.Features, "l") {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		id := q.Get("id")
		if id != "" {
			err = b.LogCon(id)
			if err != nil {
				fmt.Println(err)
				writer.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			writer.WriteHeader(http.StatusBadRequest)
		}
		return
	}
	if q.Has("userCount") {
		if !strings.Contains(api.Features, "c") {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		var count int64
		count, err = b.AppUserCount()
		if err != nil {
			fmt.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "	")
		err = enc.Encode(map[string]int64{"count": count})
		if err != nil {
			fmt.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}
	if q.Has("auth") {
		if !strings.Contains(api.Features, "g") {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		u, uPres := body["username"].(string)
		p, pPres := body["password"].(string)
		if !uPres || !pPres {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		res := b.GlobalUsers.FindOne(context.TODO(), bson.M{"username": u})
		if res.Err() != nil {
			if res.Err() == mongo.ErrNoDocuments {
				enc := json.NewEncoder(writer)
				enc.SetIndent("", "	")
				enc.Encode(map[string]any{
					"uuid":    "",
					"timeout": 0,
					"token":   "",
				})
			} else {
				fmt.Println(err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		var usr GlobalUser
		err = res.Decode(&usr)
		if err != nil {
			fmt.Println(err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if usr.Failed > 3 {
			dur := time.Minute * time.Duration(3^((usr.Failed/3)-1))
			if usr.LastTimeout == 0 {
				_, err = b.GlobalUsers.UpdateOne(context.TODO(), bson.M{"username": u}, bson.M{"lastTimeout": time.Now().Unix()})
				if err != nil {
					fmt.Println(err)
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}
				enc := json.NewEncoder(writer)
				enc.SetIndent("", "	")
				enc.Encode(map[string]any{
					"uuid":    "",
					"timeout": int(dur.Seconds()),
					"token":   "",
				})
				return
			}
			last := time.Unix(usr.LastTimeout, 0)
			if last.Add(dur).After(time.Now()) {
				enc := json.NewEncoder(writer)
				enc.SetIndent("", "	")
				enc.Encode(map[string]any{
					"uuid":    "",
					"timeout": last.Add(dur).Unix() - time.Now().Unix(),
					"token":   "",
				})
				return
			}
		}
		hashP := argon2.IDKey([]byte(p), []byte(usr.Salt), 3, 32*1024, 4, 32)
		if string(hashP) != usr.Password {
			usr.Failed++
			if usr.Failed%3 == 0 {
				usr.LastTimeout = time.Now().Unix()
			}
			_, err = b.GlobalUsers.UpdateOne(context.TODO(), bson.M{"username": u}, bson.M{"lastTimeout": usr.LastTimeout, "failed": usr.Failed})
			if err != nil {
				fmt.Println(err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			timeout := 0
			if usr.Failed%3 == 0 {
				timeout = int((time.Minute * time.Duration(3^((usr.Failed/3)-1))).Seconds())
			}
			enc := json.NewEncoder(writer)
			enc.SetIndent("", "	")
			enc.Encode(map[string]any{
				"uuid":    "",
				"timeout": timeout,
				"token":   "",
			})
			return
		}
		//TODO: generate JWT token to be returned.
	}
	writer.WriteHeader(http.StatusBadRequest)
}

// Amount of users for your application within the last month.
func (b Backend) AppUserCount() (int64, error) {
	return b.AppUsers.EstimatedDocumentCount(context.TODO())
}

// Logs a connection for the given uuid.
func (b Backend) LogCon(uuid string) error {
	n := time.Now()
	time := n.Year()*10000 + int(n.Month())*100 + n.Day()
	res := b.AppUsers.FindOneAndUpdate(context.TODO(), bson.M{"_id": uuid}, bson.M{"lastConnected": time})
	if res.Err() == mongo.ErrNoDocuments {
		_, err := b.AppUsers.InsertOne(context.TODO(), bson.M{"_id": uuid, "hasGlobal": false, "lastConnected": time})
		if err != nil {
			return err
		}
	}
	return nil
}

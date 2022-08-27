package stupid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/argon2"
)

func (b Backend) authLogin(writer http.ResponseWriter, body map[string]any) {
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
			fmt.Println(res.Err())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	var usr GlobalUser
	err := res.Decode(&usr)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	//Check if timed-out, if so, return the timeout time to the user.
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

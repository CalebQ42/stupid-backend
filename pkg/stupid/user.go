package stupid

import "time"

type logUser struct {
	LastCon  time.Time `json:"lastCon" bson:"lastCon"`
	ID       string    `json:"id" bson:"_id"`
	Platform string    `json:"platform" bson:"platform"`
}

type User struct {
}

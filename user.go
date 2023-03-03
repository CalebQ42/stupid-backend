package stupid

import "time"

type logUser struct {
	LastCon  time.Time `json:"lastCon" bson:"lastCon"`
	ID       string    `json:"id" bson:"_id"`
	Platform string    `json:"platform" bson:"platform"`
}

type AuthdUser struct {
	ID       string `json:"id" bson:"_id"`
	Username string `json:"username" bson:"username"`
	Email    string `json:"email" bson:"email"`
}

type user struct {
	LastTimeout time.Time `json:"lastTimeout" bson:"lastTimeout"`
	ID          string    `json:"id" bson:"_id"`
	Username    string    `json:"username" bson:"username"`
	Email       string    `json:"email" bson:"email"`
	Password    string    `json:"password" bson:"password"`
	Salt        string    `json:"salt" bson:"salt"`
	Failed      int       `json:"failed" bson:"failed"`
}

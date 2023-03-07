package stupid

type logUser struct {
	ID       string `json:"id" bson:"_id"`
	Platform string `json:"platform" bson:"platform"`
	LastCon  int64  `json:"lastCon" bson:"lastCon"`
}

type AuthdUser struct {
	ID       string `json:"id" bson:"_id"`
	Username string `json:"username" bson:"username"`
	Email    string `json:"email" bson:"email"`
}

type user struct {
	ID          string `json:"id" bson:"_id"`
	Username    string `json:"username" bson:"username"`
	Email       string `json:"email" bson:"email"`
	Password    string `json:"password" bson:"password"`
	Salt        string `json:"salt" bson:"salt"`
	LastTimeout int64  `json:"lastTimeout" bson:"lastTimeout"`
	Failed      int    `json:"failed" bson:"failed"`
}

package stupid

type ApiKey struct {
	Features map[string]bool `bson:"features" json:"features"`
	Key      string          `bson:"_id" json:"_id"`
	AppID    string          `bson:"appID" json:"appID"`
	Alias    string          `bson:"alias" json:"alias"`
	Death    int             `bson:"death" json:"death"`
}

type RegUser struct {
	ID          string `bson:"_id" json:"_id"`
	Username    string `bson:"username" json:"username"`
	Password    string `bson:"password" json:"password"`
	Salt        string `bson:"salt" json:"salt"`
	Email       string `bson:"email" json:"email"`
	Failed      int    `bson:"failed" json:"failed"`
	LastTimeout int    `bson:"lastTimeout" json:"lastTimeout"`
}

type ConLog struct {
	ID       string `bson:"_id" json:"_id"`
	Platform string `bson:"plat" json:"plat"`
	LastConn string `bson:"lastConn" json:"lastConn"`
}

type IndCrash struct {
	ID       string `bson:"_id" json:"_id"`
	Err      string `bson:"err" json:"err"`
	Platform string `bson:"plat" json:"plat"`
	Stack    string `bson:"stack" json:"stack"`
}

type GroupCrash struct {
	ID        string     `bson:"_id" json:"_id"`
	Err       string     `bson:"err" json:"err"`
	FirstLine string     `bson:"first" json:"first"`
	Crashes   []IndCrash `bson:"crashes" json:"crashes"`
}
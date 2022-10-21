package stupid

type ApiKey struct {
	Features map[string]bool `bson:"features" json:"features"`
	Key      string          `bson:"_id" json:"_id"`
	AppID    string          `bson:"appID" json:"appID"`
	Alias    string          `bson:"alias" json:"alias"`
	Death    int             `bson:"death" json:"death"`
}

// A registered user
type User struct {
	ID          string `bson:"_id" json:"_id"`
	Username    string `bson:"username" json:"username"`
	Password    string `bson:"password" json:"password"`
	Salt        string `bson:"salt" json:"salt"`
	Email       string `bson:"email" json:"email"`
	Failed      int    `bson:"failed" json:"failed"`
	LastTimeout int    `bson:"lastTimeout" json:"lastTimeout"`
}

// Connection log
type ConLog struct {
	ID       string `bson:"_id" json:"_id"`
	Platform string `bson:"plat" json:"plat"`
	LastConn string `bson:"lastConn" json:"lastConn"`
}

// Individual crash report
type IndCrash struct {
	ID       string `bson:"_id" json:"_id"`
	Err      string `bson:"err" json:"err"`
	Platform string `bson:"plat" json:"plat"`
	Stack    string `bson:"stack" json:"stack"`
}

// Grouped crash report
type GroupCrash struct {
	ID        string     `bson:"_id" json:"_id"`
	Err       string     `bson:"err" json:"err"`
	FirstLine string     `bson:"first" json:"first"`
	Crashes   []IndCrash `bson:"crashes" json:"crashes"`
}
